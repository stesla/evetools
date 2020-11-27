package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

func (s *Server) ShowBrowse(w http.ResponseWriter, r *http.Request) {
	groups := make(map[string]*sde.MarketGroup, len(sde.MarketGroupRoots))
	for _, g := range sde.MarketGroupRoots {
		groups[g.Name] = g
	}
	s.renderView(w, r, "browse", nil, map[string]interface{}{
		"Groups": groups,
	})
}

func (s *Server) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		internalServerError(w, "FavoriteTypes", err)
		return
	}
	favoriteTypes := make(map[string]*sde.MarketType, len(favorites))
	for _, id := range favorites {
		t, found := sde.MarketTypes[id]
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("type %d not found", id))
			return
		}
		favoriteTypes[t.Name] = t
	}

	wallet, err := s.esi.GetWalletBalance(r.Context(), user.ActiveCharacterID)
	if err != nil {
		log.Println("API Error: GetWalletBalance:", err)
	}

	var buyTotal, sellTotal float64
	orders, err := s.esi.GetMarketOrders(r.Context(), user.ActiveCharacterID)
	if err != nil {
		log.Println("API Error: GetMarketOrders:", err)
	} else {
		for _, o := range orders {
			if o.IsBuyOrder {
				buyTotal += o.Escrow
			} else {
				sellTotal += float64(o.VolumeRemain) * o.Price
			}
		}
	}

	s.renderView(w, r, "dashboard", nil, map[string]interface{}{
		"BuyTotal":      buyTotal,
		"Favorites":     favorites,
		"FavoriteTypes": favoriteTypes,
		"SellTotal":     sellTotal,
		"WalletBalance": wallet,
	})
}

const brokerRelationsID = 3446

func calculateBrokerFee(stationID int, standings []esi.Standing, skills []esi.Skill) float64 {
	var brokerRelations float64
	for _, s := range skills {
		if s.ID == brokerRelationsID {
			brokerRelations = float64(s.ActiveLevel)
		}
	}

	station, _ := sde.Stations[stationID]
	corp, _ := sde.Corporations[station.CorpID]

	var corpStanding, factionStanding float64
	for _, s := range standings {
		if s.FromType == "npc_corp" && s.FromID == corp.ID {
			corpStanding = s.Standing
		} else if s.FromType == "faction" && s.FromID == corp.FactionID {
			factionStanding = s.Standing
		}
	}

	return 0.05 - (0.003 * brokerRelations) - (0.0003 * factionStanding) - (0.0002 * corpStanding)
}

func (s *Server) ShowGroupDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID, _ := strconv.Atoi(vars["groupID"])

	group, found := sde.MarketGroups[groupID]
	if !found {
		apiInternalServerError(w, "GetMarketGroup", fmt.Errorf("no group found for ID %d", groupID))
		return
	}

	parent, _ := sde.MarketGroups[group.ParentID]

	log.Println(len(group.Groups), len(group.Types))

	groups := make(map[string]*sde.MarketGroup, len(group.Groups))
	for _, id := range group.Groups {
		g, _ := sde.MarketGroups[id]
		groups[g.Name] = g
	}

	types := make(map[string]*sde.MarketType, len(group.Types))
	for _, id := range group.Types {
		t, _ := sde.MarketTypes[id]
		types[t.Name] = t
	}

	user := currentUser(r)
	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		internalServerError(w, "FavoriteTypes", err)
		return
	}

	data := map[string]interface{}{
		"Group":     group,
		"Parent":    parent,
		"Favorites": favorites,
	}
	if len(types) == 0 {
		data["Children"] = groups
	} else {
		data["Children"] = types
		data["HasTypes"] = true
	}
	s.renderView(w, r, "groupDetails", nil, data)
}

func (s *Server) ShowMarketOrdersCurrent(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	orders, err := s.esi.GetMarketOrders(r.Context(), user.ActiveCharacterID)
	if err != nil {
		log.Println("API Error: fetching orders:", err)
	}

	buy, sell, err := s.processOrders(orders, 0)
	if err != nil {
		internalServerError(w, "processOrders", err)
		return
	}

	s.renderView(w, r, "orders", nil, map[string]interface{}{
		"Buy":  buy,
		"Sell": sell,
	})
}

func (s *Server) ShowMarketOrdersHistory(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	orders, err := s.esi.GetMarketOrderHistory(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "fetching orders", err)
		return
	}

	buy, sell, err := s.processOrders(orders, 30*24*time.Hour)
	if err != nil {
		apiInternalServerError(w, "processOrders", err)
		return
	}

	s.renderView(w, r, "orders", nil, map[string]interface{}{
		"Buy":       buy,
		"Sell":      sell,
		"IsHistory": true,
	})
}

type displayOrder struct {
	Order   *esi.MarketOrder
	Type    *sde.MarketType
	Station *sde.Station
}

func (s *Server) processOrders(orders []*esi.MarketOrder, days time.Duration) (buy, sell []*displayOrder, _ error) {
	buy = []*displayOrder{}
	sell = []*displayOrder{}
	for _, order := range orders {
		issued, err := time.Parse("2006-01-02T15:04:05Z", order.Issued)
		if err != nil {
			return nil, nil, err
		}

		if days > 0 && time.Since(issued) > days {
			continue
		}

		expires := issued.Add(time.Duration(order.Duration) * 24 * time.Hour)
		d := time.Until(expires).Round(time.Second)
		days := d / (24 * time.Hour)
		d -= days * 24 * time.Hour
		hours := d / time.Hour
		d -= hours * time.Hour
		minutes := d / time.Minute
		d -= minutes * time.Minute
		seconds := d / time.Second

		order.TimeRemaining = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)

		t, found := sde.MarketTypes[order.TypeID]
		if !found {
			return nil, nil, fmt.Errorf("no type for id %d", order.TypeID)
		}

		s, found := sde.Stations[order.LocationID]
		if !found {
			return nil, nil, fmt.Errorf("no station for id %d", order.LocationID)
			return
		}

		do := &displayOrder{Order: order, Type: t, Station: s}
		if order.IsBuyOrder {
			buy = append(buy, do)
		} else {
			sell = append(sell, do)
		}
	}
	return
}

func (s *Server) ShowSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.FormValue("q"))
	if q == "" {
		apiError(w, fmt.Errorf("must provide query string"), http.StatusBadRequest)
		return
	}

	marketTypes := map[string]*sde.MarketType{}
	for _, t := range sde.MarketTypes {
		if strings.Contains(strings.ToLower(t.Name), q) {
			marketTypes[t.Name] = t
		}
	}

	user := currentUser(r)
	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		internalServerError(w, "FavoriteTypes", err)
		return
	}

	s.renderView(w, r, "search", nil, map[string]interface{}{
		"Favorites": favorites,
		"Types":     marketTypes,
	})
}

func (s *Server) ShowSettings(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	characters, err := s.db.GetCharactersForUser(user.ID)
	if err != nil {
		apiInternalServerError(w, "GetCharactersForUser", err)
		return
	}
	charactersByName := make(map[string]*model.Character, len(characters))
	for _, c := range characters {
		charactersByName[c.CharacterName] = c
	}

	stationA, found := sde.Stations[user.StationA]
	if !found {
		apiInternalServerError(w, "GetStation", fmt.Errorf("no station for id %d", user.StationA))
		return
	}

	stationB, found := sde.Stations[user.StationB]
	if !found {
		apiInternalServerError(w, "GetStation", fmt.Errorf("no station for id %d", user.StationB))
		return
	}

	s.renderView(w, r, "settings", nil, map[string]interface{}{
		"Characters": charactersByName,
		"StationA":   stationA,
		"StationB":   stationB,
	})
}

func (s *Server) ViewTransactions(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	txns, err := s.esi.GetWalletTransactions(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetWalletTransactions", err)
		return
	}

	maxTxnID, err := s.db.GetLatestTxnID()
	if err != nil {
		apiInternalServerError(w, "GetLatestTxnID", err)
		return
	}

	types := map[int]*sde.MarketType{}
	stations := map[int]*sde.Station{}
	for _, txn := range txns {
		if txn.TransactionID > maxTxnID {
			txn.ClientName, err = s.esi.GetCharacterName(txn.ClientID)
			if err != nil {
				apiInternalServerError(w, "GetCharacterName", err)
				return
			}
			s.db.SaveTransaction(txn)
		}

		t, found := sde.MarketTypes[txn.TypeID]
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("no type for id %d", txn.TypeID))
			return
		}
		types[t.ID] = t

		s, found := sde.Stations[txn.LocationID]
		if !found {
			apiInternalServerError(w, "GetStation", fmt.Errorf("no station for id %d", txn.LocationID))
			return
		}
		stations[s.ID] = s

	}

	txns, err = s.db.GetTransactions()
	if err != nil {
		apiInternalServerError(w, "GetTransactions", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"stations":     stations,
		"transactions": txns,
		"types":        types,
	})
}

func (s *Server) ShowTypeDetails(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	marketType, found := sde.MarketTypes[typeID]
	if !found {
		apiInternalServerError(w, "GetMarketType", fmt.Errorf("no type for id %d", typeID))
		return
	}

	marketGroup, _ := sde.MarketGroups[marketType.MarketGroupID]

	parentGroups := []*sde.MarketGroup{}
	parentID := marketGroup.ParentID
	for parentID != 0 {
		g, _ := sde.MarketGroups[parentID]
		parentID = g.ParentID
		parentGroups = append(parentGroups, g)
	}

	favorite, err := s.db.IsFavorite(user.ID, typeID)
	if err != nil && err != model.ErrNotFound {
		internalServerError(w, "GetType", err)
		return
	}

	infoA, err := s.stationInfo(r.Context(), user.ActiveCharacterID, typeID, user.StationA)
	if err != nil {
		log.Println("API Error:", err)
	}

	infoB, err := s.stationInfo(r.Context(), user.ActiveCharacterID, typeID, user.StationB)
	if err != nil {
		log.Println("API Error:", err)
	}

	s.renderView(w, r, "typeDetails", nil, map[string]interface{}{
		"IsFavorite":   favorite,
		"Type":         marketType,
		"Group":        marketGroup,
		"ParentGroups": parentGroups,
		"InfoA":        infoA,
		"InfoB":        infoB,
	})
}

func (s *Server) stationInfo(ctx context.Context, characterID, typeID, stationID int) (map[string]interface{}, error) {
	station, found := sde.Stations[stationID]
	if !found {
		return nil, fmt.Errorf("no station for id %d", stationID)
	}

	solarSystem, _ := sde.SolarSystems[station.SystemID]

	skills, err := s.esi.GetCharacterSkills(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("GetCharacterSkills: %v", err)
	}

	standings, err := s.esi.GetCharacterStandings(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("GetCharacterStandings: %v", err)
	}

	price, err := s.esi.GetMarketPrices(ctx, station.ID, solarSystem.RegionID, typeID)
	if err != nil {
		return nil, fmt.Errorf("GetMarketPrices: %v", err)
	}

	history, err := s.esi.GetPriceHistory(ctx, solarSystem.RegionID, typeID)
	if err != nil {
		return nil, fmt.Errorf("GetPriceHistory: %v", err)
	}

	var volume int64
	var lowest, average, highest float64
	for _, day := range history {
		if lowest == 0 || lowest > day.Lowest {
			lowest = day.Lowest
		}
		if day.Highest > highest {
			highest = day.Highest
		}
		average += day.Average
		volume += day.Volume
	}
	if l := len(history); l > 0 {
		average /= float64(l)
		volume /= int64(l)
	}

	return map[string]interface{}{
		"average":   average,
		"brokerFee": calculateBrokerFee(station.ID, standings, skills),
		"buy":       price.Buy,
		"highest":   highest,
		"history":   history,
		"lowest":    lowest,
		"margin":    price.Margin(),
		"sell":      price.Sell,
		"station":   station,
		"system":    solarSystem,
		"volume":    volume,
	}, nil
}
