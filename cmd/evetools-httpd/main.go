package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

type contextKey int

const (
	CurrentUserKey contextKey = 1 + iota
	CurrentSessionKey
)

var store *sessions.CookieStore

func init() {
	pflag.String("addr", ":8080", "address to listen on")
	pflag.String("dir", "./public", "directory to serve files from")

	viper.BindPFlag("httpd.address", pflag.Lookup("addr"))
	viper.BindPFlag("httpd.static.dir", pflag.Lookup("dir"))
	viper.SetDefault("httpd.session.auth_key", securecookie.GenerateRandomKey(64))
	viper.SetDefault("httpd.session.name", "evetools")
	viper.SetDefault("model.database", "./evetools.sqlite3")
	viper.SetDefault("sde.database", "./eve-sde.sqlite3")
	viper.SetDefault("oauth.basePath", "https://login.eveonline.com")

	gob.Register(oauth2.Token{})
	gob.Register(model.User{})
}

var oauthConfig = oauth2.Config{
	Scopes: []string{
		"esi-ui.open_window.v1",
		"publicData",
	},
}

func main() {
	pflag.Parse()

	viper.SetConfigName(".evetools")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error loading config file: %s", err)
	}

	db, err := model.Initialize(viper.GetString("model.database"))
	if err != nil {
		log.Fatalf("error initializing model: %s", err)
	}

	static, err := sde.Initialize(viper.GetString("sde.database"))
	if err != nil {
		log.Fatalf("error initializing SDE: %s", err)
	}

	if err := initOAuthConfig(); err != nil {
		log.Fatalf("error initializing oauth: %s", err)
	}

	store = sessions.NewCookieStore([]byte(viper.GetString("httpd.session.auth_key")))

	staticFiles := http.Dir(viper.GetString("httpd.static.dir"))
	var handler http.Handler = NewServer(http.FileServer(staticFiles), db, static)
	handler = handlers.LoggingHandler(os.Stdout, handler)
	handler = handlers.ProxyHeaders(handler)

	addr := viper.GetString("httpd.address")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func initOAuthConfig() error {
	oauthConfig.ClientID = viper.GetString("oauth.clientID")
	if oauthConfig.ClientID == "" {
		return fmt.Errorf("must provide oauth.clientID")
	}
	oauthConfig.ClientSecret = viper.GetString("oauth.clientSecret")
	if oauthConfig.ClientSecret == "" {
		return fmt.Errorf("must provide oauth.clientSecret")
	}
	oauthConfig.RedirectURL = viper.GetString("oauth.redirectURL")
	if oauthConfig.RedirectURL == "" {
		return fmt.Errorf("must provide oauth.redirectURL")
	}
	basePath := viper.GetString("oauth.basePath")
	oauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/v2/oauth/authorize", basePath),
		TokenURL: fmt.Sprintf("%s/v2/oauth/token", basePath),
	}
	return nil
}

type Server struct {
	http   http.Client
	esi    *esi.Client
	mux    *mux.Router
	db     model.DB
	static sde.DB
}

func NewServer(static http.Handler, db model.DB, sdb sde.DB) *Server {
	s := &Server{
		mux:    mux.NewRouter(),
		db:     db,
		static: sdb,
	}
	s.esi = esi.NewClient(&s.http)

	s.mux.NotFoundHandler = onlyAllowGet(alwaysThisPath("/", static))
	s.mux.PathPrefix("/css").Handler(static)
	s.mux.PathPrefix("/data").Handler(static)
	s.mux.PathPrefix("/js").Handler(static)
	s.mux.PathPrefix("/views").Handler(static)
	s.mux.Methods("GET").Path("/login").HandlerFunc(s.Login)
	s.mux.Methods("GET").Path("/login/callback").HandlerFunc(s.LoginCallback)
	s.mux.Methods("GET").Path("/logout").HandlerFunc(s.Logout)

	api := s.mux.PathPrefix("/api").Subrouter()
	api.Use(haveLoggedInUser)
	api.Use(contentType("application/json").Middleware)
	api.Methods("GET").Path("/v1/currentUser").HandlerFunc(s.GetCurrentUser)
	api.Methods("GET").Path("/v1/stations").HandlerFunc(s.GetStations)
	api.Methods("GET").Path("/v1/types/search/{filter}").HandlerFunc(s.GetTypeSearch)
	api.Methods("GET").Path("/v1/types/{typeID:[0-9]+}").HandlerFunc(s.GetTypeID)
	api.Methods("PUT").Path("/v1/types/{typeID:[0-9]+}/favorite").HandlerFunc(s.PutTypeFavorite)
	api.Methods("POST").Path("/v1/types/{typeID:[0-9]+}/openInGame").HandlerFunc(s.PostOpenInGame)
	api.Methods("PUT").Path("/v1/user/station").HandlerFunc(s.PutUserStation)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, &s.http)
	s.mux.ServeHTTP(w, r.WithContext(ctx))
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, viper.GetString("httpd.session.name"))
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.RawURLEncoding.EncodeToString(b)
	session.Values["oauth.state"] = state
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "session.Save", err)
		return
	}
	url := oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) LoginCallback(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		return
	}

	state := r.FormValue("state")
	if str, ok := session.Values["oauth.state"].(string); !ok || str != state {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		internalServerError(w, "oauth.Exchange", err)
		return
	}

	resp, err := s.http.Get(fmt.Sprintf("%s/.well-known/oauth-authorization-server",
		viper.GetString("oauth.basePath")))
	if err != nil {
		internalServerError(w, "GET metadata", err)
		return
	}
	defer resp.Body.Close()
	var meta struct {
		URI string `json:"jwks_uri"`
	}
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		internalServerError(w, "Decode metadata", err)
		return
	}

	keyset, err := jwk.Fetch(meta.URI, jwk.WithHTTPClient(&s.http))
	if err != nil {
		internalServerError(w, "fetch jwks", err)
		return
	}

	payload, err := jwt.ParseString(token.AccessToken, jwt.WithKeySet(keyset))
	if err != nil {
		internalServerError(w, "verify token", err)
		return
	}

	chunks := strings.Split(payload.Subject(), ":")
	if len(chunks) != 3 {
		err := fmt.Errorf("incorrect subject format %q", payload.Subject())
		internalServerError(w, "get characterID", err)
		return
	}
	characterID, err := strconv.Atoi(chunks[2])
	if err != nil {
		internalServerError(w, "get characterID", err)
		return
	}

	v, _ := payload.Get("name")
	characterName, ok := v.(string)
	if !ok {
		internalServerError(w, "get characterName", err)
		return
	}

	v, _ = payload.Get("owner")
	ownerToken, ok := v.(string)
	if !ok {
		internalServerError(w, "get owner token", err)
		return
	}

	user, err := s.db.FindOrCreateUserForCharacter(characterID, characterName, ownerToken)
	if err != nil {
		internalServerError(w, "FindOrCreateUserForCharacter", err)
		return
	}

	session.Values["user"] = user
	session.Values["token"] = token
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "save session", err)
		return
	}

	next := "/"
	if str, ok := session.Values["next"].(string); ok && str != "" {
		next = str
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		return
	}
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}
