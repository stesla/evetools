package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

func (s *Server) DeleteUserCharacter(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)

	vars := mux.Vars(r)
	characterID, _ := strconv.Atoi(vars["characterID"])

	err := s.db.RemoveCharacterForUser(user.ID, characterID)
	if err != nil {
		apiInternalServerError(w, "RemoveCharacterForUser", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetStations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(strings.ToLower(r.FormValue("q")))
	result := []*sde.Station{}
	for _, s := range sde.Stations {
		if name := strings.ToLower(s.Name); strings.HasPrefix(name, query) {
			result = append(result, s)
		}
	}
	json.NewEncoder(w).Encode(result)
}

func (s *Server) GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	favorites, err := s.db.GetFavoriteTypes(user.ID)
	if err != nil {
		apiInternalServerError(w, "FavoriteTypes", err)
		return
	}
	json.NewEncoder(w).Encode(&favorites)
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

func (s *Server) PutUserStationA(w http.ResponseWriter, r *http.Request) {
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
	err = s.db.SaveUserStationA(user.ID, station.ID)
	if err != nil {
		apiInternalServerError(w, "SaveUserStation", err)
		return
	}

	session := currentSession(r)
	user.StationA = station.ID
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		apiInternalServerError(w, "save session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) PutUserStationB(w http.ResponseWriter, r *http.Request) {
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
	err = s.db.SaveUserStationB(user.ID, station.ID)
	if err != nil {
		apiInternalServerError(w, "SaveUserStation", err)
		return
	}

	session := currentSession(r)
	user.StationA = station.ID
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		apiInternalServerError(w, "save session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
