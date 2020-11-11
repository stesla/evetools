package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	character, err := s.db.GetCharacter(user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetCharacter", err)
		return
	}

	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		apiInternalServerError(w, "GetStationByID", err)
		return
	}

	favorites, err := s.db.FavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": character,
		"favorites": favorites,
		"station":   station,
	})
}

func (s *Server) GetStations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.FormValue("q"))
	if len(query) < 3 {
		apiError(w, errors.New("q must be at least three characters"), http.StatusBadRequest)
		return
	}

	stations, err := s.static.GetStations(query)
	if err != nil {
		apiInternalServerError(w, "GetStations", err)
		return
	}

	json.NewEncoder(w).Encode(&stations)
}

func (s *Server) GetTypeID(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		apiInternalServerError(w, "GetStationByID", err)
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["typeID"])

	isFavorite, err := s.db.IsFavorite(user.ID, id)
	if err != nil && err != model.ErrNotFound {
		apiInternalServerError(w, "GetType", err)
		return
	}

	price, err := s.esi.MarketPrices(r.Context(), station.ID, station.Region.ID, id)
	if err != nil {
		apiInternalServerError(w, "JitaPrices", err)
		return
	}

	history, err := s.esi.MarketHistory(r.Context(), station.Region.ID, id)
	if err != nil {
		apiInternalServerError(w, "JitaHistory", err)
		return
	}

	var volume int64
	var lowest, average, highest float64
	for _, day := range history {
		lowest += day.Lowest
		average += day.Average
		highest += day.Highest
		volume += day.Volume
	}
	if l := len(history); l > 0 {
		lowest /= float64(l)
		average /= float64(l)
		highest /= float64(l)
		volume /= int64(l)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"buy":      price.Buy,
		"sell":     price.Sell,
		"margin":   price.Margin(),
		"volume":   volume,
		"lowest":   lowest,
		"average":  average,
		"highest":  highest,
		"history":  history,
		"favorite": isFavorite,
	})
}

func (s *Server) GetTypeSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.static.SearchTypesByName(vars["filter"])
	if err != nil {
		apiInternalServerError(w, "GetMarketTypes", err)
		return
	}

	json.NewEncoder(w).Encode(items)
}

func (s *Server) PostOpenInGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	if err := s.esi.OpenMarketWindow(r.Context(), typeID); err != nil {
		apiInternalServerError(w, "OpenMarketWindow", err)
	} else {
		w.Header()["Content-Type"] = nil
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) PutTypeFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	var req struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = fmt.Errorf("error parsing request body: %v", err)
		apiError(w, err, http.StatusBadRequest)
		return
	}

	user := currentUser(r)
	if err := s.db.SetFavorite(user.ID, typeID, req.Favorite); err != nil {
		apiInternalServerError(w, "SetFavorite", err)
		return
	}

	json.NewEncoder(w).Encode(&req)
}

func (s *Server) PutUserStation(w http.ResponseWriter, r *http.Request) {
	var station sde.Station
	err := json.NewDecoder(r.Body).Decode(&station)
	if err != nil {
		err = fmt.Errorf("error parsing request body: %v", err)
		apiError(w, err, http.StatusBadRequest)
		return
	}

	user := currentUser(r)
	err = s.db.SaveUserStation(user.ID, station.ID)
	if err != nil {
		apiInternalServerError(w, "SaveUserStation", err)
		return
	}

	session := currentSession(r)
	user.StationID = station.ID
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		apiInternalServerError(w, "save session", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}