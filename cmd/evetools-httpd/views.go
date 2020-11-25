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

func (s *Server) ViewBrowse(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": sde.MarketGroupRoots,
	})
}

func (s *Server) ViewDashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}
	favoriteTypes := make([]*sde.MarketType, len(favorites))
	for i, id := range favorites {
		t, found := sde.MarketTypes[id]
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("type %d not found", id))
			return
		}
		favoriteTypes[i] = t
	}

	wallet, err := s.esi.GetWalletBalance(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "WalletBalance", err)
		return
	}

	skills, err := s.esi.GetCharacterSkills(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetCharacterSkills", err)
		return
	}

	standings, err := s.esi.GetCharacterStandings(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetCharacterStandings", err)
		return
	}

	var buyTotal, sellTotal float64
	orders, err := s.esi.GetMarketOrders(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetMarketOrders", err)
		return
	}
	for _, o := range orders {
		if o.IsBuyOrder {
			buyTotal += o.Escrow
		} else {
			sellTotal += float64(o.VolumeRemain) * o.Price
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"broker_fee":     calculateBrokerFee(user.StationA, standings, skills),
		"buy_total":      buyTotal,
		"favorites":      favoriteTypes,
		"sell_total":     sellTotal,
		"wallet_balance": wallet,
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

func (s *Server) ViewGroupDetails(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	vars := mux.Vars(r)
	groupID, _ := strconv.Atoi(vars["groupID"])

	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}

	group, found := sde.MarketGroups[groupID]
	if !found {
		apiInternalServerError(w, "GetMarketGroup", fmt.Errorf("no group found for ID %d", groupID))
		return
	}

	parent, _ := sde.MarketGroups[group.ParentID]

	log.Println(len(group.Groups), len(group.Types))

	groups := make([]*sde.MarketGroup, len(group.Groups))
	for i, id := range group.Groups {
		g, _ := sde.MarketGroups[id]
		groups[i] = g
	}

	types := make([]*sde.MarketType, len(group.Types))
	for i, id := range group.Types {
		t, _ := sde.MarketTypes[id]
		types[i] = t
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"favorites": favorites,
		"group":     group,
		"groups":    groups,
		"parent":    parent,
		"types":     types,
	})
}

func (s *Server) ViewMarketOrders(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	var days time.Duration
	if str := r.FormValue("days"); str != "" {
		if i, err := strconv.Atoi(str); err != nil {
			apiError(w, fmt.Errorf("'days' must be an integer"), http.StatusBadRequest)
			return
		} else {
			days = time.Duration(i)
		}
	}

	var f func(context.Context, int) ([]*esi.MarketOrder, error)
	if days > 0 {
		f = s.esi.GetMarketOrderHistory
	} else {
		f = s.esi.GetMarketOrders
	}

	orders, err := f(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "fetching orders", err)
		return
	}

	buy, sell, err := s.processOrders(orders, days*24*time.Hour)
	if err != nil {
		apiInternalServerError(w, "processOrders", err)
		return
	}

	types := map[int]*sde.MarketType{}
	stations := map[int]*sde.Station{}
	for _, o := range orders {
		t, found := sde.MarketTypes[o.TypeID]
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("no type for id %d", o.TypeID))
			return
		}
		types[t.ID] = t

		s, found := sde.Stations[o.LocationID]
		if !found {
			apiInternalServerError(w, "GetStation", fmt.Errorf("no station for id %d", o.LocationID))
			return
		}
		stations[s.ID] = s
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"buy":      buy,
		"sell":     sell,
		"stations": stations,
		"types":    types,
	})
}

func (s *Server) processOrders(orders []*esi.MarketOrder, days time.Duration) (buy, sell []*esi.MarketOrder, _ error) {
	buy = []*esi.MarketOrder{}
	sell = []*esi.MarketOrder{}
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

		if order.IsBuyOrder {
			buy = append(buy, order)
		} else {
			sell = append(sell, order)
		}
	}
	return
}

func (s *Server) ViewSearch(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	q := strings.ToLower(r.FormValue("q"))
	if q == "" {
		apiError(w, fmt.Errorf("must provide query string"), http.StatusBadRequest)
		return
	}

	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}

	marketTypes := []*sde.MarketType{}
	for _, t := range sde.MarketTypes {
		if strings.Contains(strings.ToLower(t.Name), q) {
			marketTypes = append(marketTypes, t)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"favorites": favorites,
		"types":     marketTypes,
	})
}

func (s *Server) ViewSettings(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	characters, err := s.db.GetCharactersForUser(user.ID)
	if err != nil {
		apiInternalServerError(w, "GetCharactersForUser", err)
		return
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"characters": characters,
		"stationA":   stationA,
		"stationB":   stationB,
		"stations":   sde.Stations,
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

func (s *Server) ViewTypeDetails(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	marketType, found := sde.MarketTypes[typeID]
	if !found {
		apiInternalServerError(w, "GetMarketType", fmt.Errorf("no type for id %d", typeID))
		return
	}

	group, _ := sde.MarketGroups[marketType.MarketGroupID]

	parentGroups := []*sde.MarketGroup{}
	parentID := group.ParentID
	for parentID != 0 {
		g, _ := sde.MarketGroups[parentID]
		parentID = g.ParentID
		parentGroups = append(parentGroups, g)
	}

	favorite, err := s.db.IsFavorite(user.ID, typeID)
	if err != nil && err != model.ErrNotFound {
		apiInternalServerError(w, "GetType", err)
		return
	}

	station, found := sde.Stations[user.StationA]
	if !found {
		apiInternalServerError(w, "GetStation", fmt.Errorf("no station for id %d", user.StationA))
		return
	}

	solarSystem, _ := sde.SolarSystems[station.SystemID]

	price, err := s.esi.GetMarketPrices(r.Context(), station.ID, solarSystem.RegionID, typeID)
	if err != nil {
		apiInternalServerError(w, "JitaPrices", err)
		return
	}

	history, err := s.esi.GetPriceHistory(r.Context(), solarSystem.RegionID, typeID)
	if err != nil {
		apiInternalServerError(w, "JitaHistory", err)
		return
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"favorite": favorite,
		"marketInfo": map[string]interface{}{
			"average": average,
			"buy":     price.Buy,
			"highest": highest,
			"history": history,
			"lowest":  lowest,
			"margin":  price.Margin(),
			"sell":    price.Sell,
			"volume":  volume,
		},
		"group":         group,
		"parent_groups": parentGroups,
		"station":       station,
		"system":        solarSystem,
		"type":          marketType,
	})
}
