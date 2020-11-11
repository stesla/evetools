package main

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
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

func internalServerError(w http.ResponseWriter, tag, body string, err error) {
	log.Println("Internal Server Error:", tag, ":", err)
	http.Error(w, body, http.StatusInternalServerError)
}

func currentSession(r *http.Request) *sessions.Session {
	return r.Context().Value(CurrentSessionKey).(*sessions.Session)
}

func currentUser(r *http.Request) model.User {
	return r.Context().Value(CurrentUserKey).(model.User)
}

func haveLoggedInUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, viper.GetString("httpd.session.name"))
		if err != nil {
			internalServerError(w, "store.Get", "{}", err)
			return
		}

		user, ok := session.Values["user"].(model.User)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		oldTok, ok := session.Values["token"].(oauth2.Token)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokSrc := oauthConfig.TokenSource(r.Context(), &oldTok)
		newTok, err := tokSrc.Token()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CurrentSessionKey, session)
		ctx = context.WithValue(ctx, CurrentUserKey, user)
		ctx = context.WithValue(ctx, ESITokenKey, newTok.AccessToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
