package main

import (
	"encoding/json"
	"fmt"
	"net/http"

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
