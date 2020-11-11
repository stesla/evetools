package main

import (
	"encoding/json"
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
		internalServerError(w, "GetCharacter", "{}", err)
		return
	}

	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		internalServerError(w, "GetStationByID", "{}", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": character,
		"station":   station,
	})
}

func (s *Server) GetFavorites(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	types, err := s.db.FavoriteTypes(user.ID)
	if err != nil {
		internalServerError(w, "FavoriteTypes", "{}", err)
		return
	}

	json.NewEncoder(w).Encode(&types)
}

func (s *Server) GetStations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.FormValue("q"))
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stations, err := s.static.GetStations(query)
	if err != nil {
		internalServerError(w, "GetStations", "{}", err)
		return
	}

	json.NewEncoder(w).Encode(&stations)
}

func (s *Server) GetTypeID(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		internalServerError(w, "GetStationByID", "{}", err)
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["typeID"])

	isFavorite, err := s.db.IsFavorite(user.ID, id)
	if err != nil && err != model.ErrNotFound {
		internalServerError(w, "GetType", "{}", err)
		return
	}

	price, err := s.esi.MarketPrices(r.Context(), station.ID, station.Region.ID, id)
	if err != nil {
		internalServerError(w, "JitaPrices", "{}", err)
		return
	}

	history, err := s.esi.MarketHistory(r.Context(), station.Region.ID, id)
	if err != nil {
		internalServerError(w, "JitaHistory", "{}", err)
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
		internalServerError(w, "GetMarketTypes", "{}", err)
		return
	}

	json.NewEncoder(w).Encode(items)
}

func (s *Server) PostOpenInGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	if err := s.esi.OpenMarketWindow(r.Context(), typeID); err != nil {
		internalServerError(w, "OpenMarketWindow", "{}", err)
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := currentUser(r)
	if err := s.db.SetFavorite(user.ID, typeID, req.Favorite); err != nil {
		internalServerError(w, "SetFavorite", "{}", err)
		return
	}

	json.NewEncoder(w).Encode(&req)
}

func (s *Server) PutUserStation(w http.ResponseWriter, r *http.Request) {
	var station sde.Station
	err := json.NewDecoder(r.Body).Decode(&station)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := currentUser(r)
	err = s.db.SaveUserStation(user.ID, station.ID)
	if err != nil {
		internalServerError(w, "SaveUserStation", "{}", err)
		return
	}

	session := currentSession(r)
	user.StationID = station.ID
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "save session", "{}", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
