package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/sde"
)

func (s *Server) ViewDashboard(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}
	favoriteTypes := make([]*sde.MarketType, len(favorites))
	for i, id := range favorites {
		t, found := sde.GetMarketType(id)
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("type %d not found", id))
			return
		}
		var f sde.MarketType = *t
		f.Description = ""
		favoriteTypes[i] = &f
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
		"broker_fee":     calculateBrokerFee(user.StationID, standings, skills),
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

	station, _ := sde.GetStation(stationID)
	corp, _ := sde.GetCorporation(station.CorpID)

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
		t, found := sde.GetMarketType(o.TypeID)
		if !found {
			apiInternalServerError(w, "GetMarketType", fmt.Errorf("no type for id %d", o.TypeID))
			return
		}
		types[t.ID] = t

		s, found := sde.GetStation(o.LocationID)
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
