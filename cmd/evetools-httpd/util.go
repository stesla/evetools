package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/model"
	"golang.org/x/oauth2"
)

func alwaysThisPath(path string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		h.ServeHTTP(w, r2)
	})
}

func onlyAllowGet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type contentType string

func (s contentType) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", string(s))
		h.ServeHTTP(w, r)
	})
}

func internalServerError(w http.ResponseWriter, tag string, err error) {
	log.Println("Internal Server Error:", tag+":", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func apiError(w http.ResponseWriter, err error, status int) {
	// API calls are always sent with Content-Type: application/json,
	// So we want to send a JSON object always.
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func apiInternalServerError(w http.ResponseWriter, tag string, err error) {
	log.Println("Internal Server Error:", tag+":", err)
	apiError(w, err, http.StatusInternalServerError)
}

func currentSession(r *http.Request) *sessions.Session {
	return r.Context().Value(CurrentSessionKey).(*sessions.Session)
}

func currentUser(r *http.Request) *model.User {
	return r.Context().Value(CurrentUserKey).(*model.User)
}

func getSession(r *http.Request) *sessions.Session {
	return r.Context().Value(CurrentSessionKey).(*sessions.Session)
}

func (s *Server) haveSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()

		session, err := store.Get(r, viper.GetString("httpd.session.name"))
		if err != nil {
			internalServerError(w, "store.Get", err)
			return
		}
		ctx = context.WithValue(ctx, CurrentSessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) haveLoggedInUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/login") || strings.HasPrefix(r.URL.Path, "/logout") {
			next.ServeHTTP(w, r)
		}

		var ctx = r.Context()
		session := getSession(r)
		login := func() {
			if r.URL.Path == "/" {
				s.renderView(w, r, "login", nil, nil)
			} else {
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}

		oldTok, ok := session.Values["token"].(oauth2.Token)
		if !ok {
			login()
			return
		}
		tokSrc := oauthConfig.TokenSource(r.Context(), &oldTok)
		newTok, err := tokSrc.Token()
		if err != nil {
			login()
			return
		}
		ctx = context.WithValue(ctx, esi.AccessTokenKey, newTok.AccessToken)

		sessionUserID, ok := session.Values["userid"].(int)
		if ok {
			user, err := s.db.GetUser(sessionUserID)
			if err != nil {
				internalServerError(w, "GetUser", err)
				return
			}
			ctx = context.WithValue(ctx, CurrentUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			login()
		}
	})
}

func refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	oldTok := oauth2.Token{RefreshToken: refreshToken}
	tokSrc := oauthConfig.TokenSource(ctx, &oldTok)
	token, err := tokSrc.Token()
	if err != nil {
		return nil, err
	}
	return token, nil
}
