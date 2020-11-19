package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

func (s *Server) GetType(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["typeID"])

	marketType, found := sde.GetMarketType(id)
	if !found {
		apiError(w, fmt.Errorf("not found"), http.StatusNotFound)
		return
	}

	locationID, err := strconv.Atoi(r.FormValue("location_id"))
	if err != nil {
		apiError(w, fmt.Errorf("location_id must be an integer"), http.StatusBadRequest)
		return
	}

	regionID, err := strconv.Atoi(r.FormValue("region_id"))
	if err != nil {
		apiError(w, fmt.Errorf("region_id must be an integer"), http.StatusBadRequest)
		return
	}

	isFavorite, err := s.db.IsFavorite(user.ID, id)
	if err != nil && err != model.ErrNotFound {
		apiInternalServerError(w, "GetType", err)
		return
	}

	price, err := s.esi.GetMarketPrices(r.Context(), locationID, regionID, id)
	if err != nil {
		apiInternalServerError(w, "JitaPrices", err)
		return
	}

	history, err := s.esi.GetPriceHistory(r.Context(), regionID, id)
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
		"type":     marketType,
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

func (s *Server) GetUserCharacters(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	characters, err := s.db.GetCharactersForUser(user.ID)
	if err != nil {
		apiInternalServerError(w, "GetCharactersForUser", err)
		return
	}

	json.NewEncoder(w).Encode(characters)
}

func (s *Server) GetUserCurrent(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	character, err := s.db.GetCharacterByOwnerHash(user.ActiveCharacterHash)
	if err != nil {
		apiInternalServerError(w, "GetCharacterByOwnerHash", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"character":  character,
		"station_id": user.StationID,
	})
}

func (s *Server) GetUserSkills(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	skills, err := s.esi.GetCharacterSkills(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetCharacterSkills", err)
		return
	}

	json.NewEncoder(w).Encode(skills)
}

func (s *Server) GetUserStandings(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	standings, err := s.esi.GetCharacterStandings(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "GetCharacterStandings", err)
		return
	}

	json.NewEncoder(w).Encode(standings)
}

func (s *Server) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
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

	for _, txn := range txns {
		if txn.TransactionID > maxTxnID {
			txn.ClientName, err = s.esi.GetCharacterName(txn.ClientID)
			if err != nil {
				apiInternalServerError(w, "GetCharacterName", err)
				return
			}
			s.db.SaveTransaction(txn)
		}
	}

	txns, err = s.db.GetTransactions()
	if err != nil {
		apiInternalServerError(w, "GetTransactions", err)
		return
	}

	json.NewEncoder(w).Encode(txns)
}

func (s *Server) GetUserWalletBalance(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	wallet, err := s.esi.GetWalletBalance(r.Context(), user.ActiveCharacterID)
	if err != nil {
		apiInternalServerError(w, "WalletBalance", err)
		return
	}

	json.NewEncoder(w).Encode(wallet)
}

func (s *Server) GetVerify(w http.ResponseWriter, r *http.Request) {
	verify, err := s.esi.Verify(r.Context())
	if err != nil {
		apiError(w, fmt.Errorf("not authorized"), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"character_id":   verify.CharacterID,
		"character_name": verify.CharacterName,
	})
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

func (s *Server) PostUserCharacterActivate(w http.ResponseWriter, r *http.Request) {
	session := currentSession(r)
	user := currentUser(r)

	vars := mux.Vars(r)
	characterID, _ := strconv.Atoi(vars["characterID"])

	character, err := s.db.GetCharacterByUserAndCharacterID(user.ID, characterID)
	if err == model.ErrNotFound {
		apiError(w, fmt.Errorf("not found"), http.StatusNotFound)
		return
	} else if err != nil {
		apiInternalServerError(w, "GetCharacter", err)
		return
	}
	user.ActiveCharacterID = character.CharacterID
	user.ActiveCharacterHash = character.CharacterOwnerHash

	token, err := s.db.GetTokenForCharacter(character.ID)
	if err != nil {
		apiInternalServerError(w, "GetTokenForCharacter", err)
		return
	}
	jwt, err := refreshToken(r.Context(), token.RefreshToken)
	if err != nil {
		apiInternalServerError(w, "refreshToken", err)
		return
	}

	err = s.db.SaveActiveCharacterHash(user.ID, character.CharacterOwnerHash)
	if err != nil {
		apiInternalServerError(w, "SaveActiveCharacterHash", err)
		return
	}

	session.Values["token"] = jwt
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "session.Save", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
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
	var station struct {
		ID int `json:"id"`
	}
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

	w.WriteHeader(http.StatusNoContent)
}
