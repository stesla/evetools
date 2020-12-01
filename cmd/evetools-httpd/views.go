package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

func (s *Server) PostUserFavorites(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	names := strings.Split(r.FormValue("items"), "\n")
	for _, name := range names {
		name = strings.TrimSpace(name)
		t, found := sde.MarketTypesByName[name]
		if found {
			s.db.SetFavorite(user.ID, t.ID, true)
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

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
	var myPrices = map[int]esi.Price{}
	orders, err := s.esi.GetCharacterOrders(r.Context(), user.ActiveCharacterID)
	if err != nil {
		log.Println("API Error: GetCharacterOrders:", err)
	} else {
		for _, o := range orders {
			if o.IsBuyOrder {
				buyTotal += o.Escrow
			} else {
				sellTotal += float64(o.VolumeRemain) * o.Price
			}

			price := myPrices[o.TypeID]
			if o.IsBuyOrder && o.Price > price.Buy {
				price.Buy = o.Price
			} else if price.Sell == 0 || o.Price < price.Sell {
				price.Sell = o.Price
			}
			myPrices[o.TypeID] = price
		}
	}

	stationA := sde.Stations[user.StationA]
	systemA := sde.SolarSystems[stationA.SystemID]
	stationAPrices, err := s.db.GetPricesForStation(stationA.ID)
	if err != nil {
		internalServerError(w, "GetPricesForStation", err)
		return
	}

	stationB := sde.Stations[user.StationB]
	systemB := sde.SolarSystems[stationB.SystemID]
	stationBPrices, err := s.db.GetPricesForStation(stationB.ID)
	if err != nil {
		internalServerError(w, "GetPricesForStation", err)
		return
	}

	funcs := template.FuncMap{
		"marginAtoB": func(t *sde.MarketType) float64 {
			buy := stationAPrices[t.ID].Buy
			if buy == 0 {
				return 0
			}
			sell := stationBPrices[t.ID].Sell
			return (sell - buy) / buy
		},
		"myBuy":        func(t *sde.MarketType) float64 { return myPrices[t.ID].Buy },
		"mySell":       func(t *sde.MarketType) float64 { return myPrices[t.ID].Sell },
		"stationABuy":  func(t *sde.MarketType) float64 { return stationAPrices[t.ID].Buy },
		"stationASell": func(t *sde.MarketType) float64 { return stationAPrices[t.ID].Sell },
		"stationBBuy":  func(t *sde.MarketType) float64 { return stationBPrices[t.ID].Buy },
		"stationBSell": func(t *sde.MarketType) float64 { return stationBPrices[t.ID].Sell },
		"systemA":      func() string { return systemA.Name },
		"systemB":      func() string { return systemB.Name },
	}

	s.renderView(w, r, "dashboard", funcs, map[string]interface{}{
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

	orders, err := s.esi.GetCharacterOrders(r.Context(), user.ActiveCharacterID)
	if err != nil {
		log.Println("API Error: fetching orders:", err)
	}

	buy, sell, err := s.processOrders(orders, 0)
	if err != nil {
		internalServerError(w, "processOrders", err)
		return
	}

	buyByStation := sortByStation(buy)
	sellByStation := sortByStation(sell)

	s.renderView(w, r, "ordersCurrent", nil, map[string]interface{}{
		"Buy":  buyByStation,
		"Sell": sellByStation,
	})
}

func sortByStation(orders []*displayOrder) map[string][]*displayOrder {
	sort.Sort(byTypeName{orders})
	result := map[string][]*displayOrder{}
	for _, o := range orders {
		s, ok := result[o.Station.Name]
		if !ok {
			s = []*displayOrder{}
		}
		s = append(s, o)
		result[o.Station.Name] = s
	}
	return result
}

func (s *Server) ShowMarketOrdersHistory(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	orders, err := s.esi.GetCharacterOrderHistory(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "fetching orders", err)
		return
	}

	buy, sell, err := s.processOrders(orders, 30*24*time.Hour)
	if err != nil {
		apiInternalServerError(w, "processOrders", err)
		return
	}

	sort.Sort(sort.Reverse(byOrderID{buy}))
	sort.Sort(sort.Reverse(byOrderID{sell}))

	s.renderView(w, r, "ordersHistory", nil, map[string]interface{}{
		"Buy":  buy,
		"Sell": sell,
	})
}

type displayOrder struct {
	Order   *esi.MarketOrder
	Type    *sde.MarketType
	Station *sde.Station
}

type displayOrders []*displayOrder

func (s displayOrders) Len() int      { return len(s) }
func (s displayOrders) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type byOrderID struct{ displayOrders }

func (s byOrderID) Less(i, j int) bool {
	return s.displayOrders[i].Order.OrderID < s.displayOrders[j].Order.OrderID
}

type byTypeName struct{ displayOrders }

func (s byTypeName) Less(i, j int) bool {
	return s.displayOrders[i].Type.Name < s.displayOrders[j].Type.Name
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

func (s *Server) ShowTransactions(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	txns, err := s.esi.GetWalletTransactions(r.Context(), user.ActiveCharacterID)
	if err != nil {
		internalServerError(w, "GetWalletTransactions", err)
		return
	}

	clientIDs := make([]int, len(txns))
	for i, txn := range txns {
		clientIDs[i] = txn.ClientID
	}
	characterNames, err := s.esi.GetNames(clientIDs)
	if err != nil {
		internalServerError(w, "GetNames", err)
		return
	}

	funcs := template.FuncMap{
		"character": func(txn *esi.WalletTransaction) string {
			if name, found := characterNames[txn.ClientID]; found {
				return name
			} else {
				return "UKNOWN"
			}
		},
		"station": func(txn *esi.WalletTransaction) (string, error) {
			station, ok := sde.Stations[txn.LocationID]
			if !ok {
				return "", fmt.Errorf("station not found for id %d", txn.LocationID)
			}
			return station.Name, nil
		},
		"total": func(txn *esi.WalletTransaction) float64 {
			return float64(txn.Quantity) * txn.UnitPrice
		},
		"type": func(txn *esi.WalletTransaction) (*sde.MarketType, error) {
			t, ok := sde.MarketTypes[txn.TypeID]
			if !ok {
				return nil, fmt.Errorf("type not found for id %d", txn.TypeID)
			}
			return t, nil
		},
	}
	s.renderView(w, r, "transactions", funcs, map[string]interface{}{
		"Transactions": txns,
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

	parentGroupsReverse := []*sde.MarketGroup{}
	parentID := marketGroup.ParentID
	for parentID != 0 {
		g, _ := sde.MarketGroups[parentID]
		parentID = g.ParentID
		parentGroupsReverse = append(parentGroupsReverse, g)
	}
	// Because we're appending to the list from the item up, the groups end up
	// in it backwards, so we need to switch them around.
	parentGroups := make([]*sde.MarketGroup, len(parentGroupsReverse))
	for i := range parentGroups {
		k := len(parentGroupsReverse) - i - 1
		parentGroups[i] = parentGroupsReverse[k]
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

	price, err := s.esi.GetMarketPriceForType(station.ID, solarSystem.RegionID, typeID)
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
