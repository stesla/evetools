package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/model"
)

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
