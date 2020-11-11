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
	"net/url"
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
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
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
	esi    *ESIClient
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
	s.esi = NewESIClient(&s.http)

	s.mux.NotFoundHandler = onlyAllowGet(alwaysThisPath("/", static))
	s.mux.PathPrefix("/css").Handler(static)
	s.mux.PathPrefix("/data").Handler(static)
	s.mux.PathPrefix("/js").Handler(static)
	s.mux.PathPrefix("/views").Handler(static)
	s.mux.Methods("GET").Path("/login").HandlerFunc(s.Login)
	s.mux.Methods("GET").Path("/login/callback").HandlerFunc(s.LoginCallback)
	s.mux.Methods("GET").Path("/logout").HandlerFunc(s.Logout)

	api := s.mux.PathPrefix("/api").Subrouter()
	api.Use(contentType("application/json").Middleware)
	api.Methods("GET").Path("/v1/currentUser").HandlerFunc(s.CurrentUser)
	api.Methods("GET").Path("/v1/stations").HandlerFunc(s.GetStations)
	api.Methods("GET").Path("/v1/types/favorites").HandlerFunc(s.TypeGetFavorites)
	api.Methods("GET").Path("/v1/types/search/{filter}").HandlerFunc(s.TypeSearch)
	api.Methods("GET").Path("/v1/types/{typeID:[0-9]+}").HandlerFunc(s.TypeDetails)
	api.Methods("PUT").Path("/v1/types/{typeID:[0-9]+}/favorite").HandlerFunc(s.TypeSetFavorite)
	api.Methods("POST").Path("/v1/types/{typeID:[0-9]+}/openInGame").HandlerFunc(s.TypeOpenInGame)
	api.Methods("PUT").Path("/v1/user/station").HandlerFunc(s.SaveUserStation)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, &s.http)
	s.mux.ServeHTTP(w, r.WithContext(ctx))
}

func (s *Server) CurrentUser(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		return
	}

	user, ok := session.Values["user"].(model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	character, err := s.db.GetCharacter(user.ActiveCharacterID)
	if err != nil {
		internalServerError(w, "GetCharacter", err)
		fmt.Fprintln(w, "{}")
		return
	}

	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		internalServerError(w, "GetStationByID", err)
		fmt.Fprintln(w, "{}")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"character": character,
		"station":   station,
	})
}

func (s *Server) GetStations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.FormValue("q"))
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{}")
		return
	}

	stations, err := s.static.GetStations(query)
	if err != nil {
		internalServerError(w, "GetStations", err)
		fmt.Fprintln(w, "{}")
		return
	}

	json.NewEncoder(w).Encode(&stations)
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

func internalServerError(w http.ResponseWriter, tag string, err error) {
	log.Println("Internal Server Error:", tag, ":", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
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
		internalServerError(w, "get characterID", fmt.Errorf("incorrect subject format %q", payload.Subject()))
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

func (s *Server) SaveUserStation(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		fmt.Fprintln(w, "{}")
		return
	}

	user, ok := session.Values["user"].(model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	var station sde.Station
	err = json.NewDecoder(r.Body).Decode(&station)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{}")
		return
	}

	err = s.db.SaveUserStation(user.ID, station.ID)
	if err != nil {
		internalServerError(w, "SaveUserStation", err)
		fmt.Fprintln(w, "{}")
		return
	}

	user.StationID = station.ID
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "save session", err)
		fmt.Fprintln(w, "{}")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) TypeDetails(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		fmt.Fprintln(w, "{}")
		return
	}

	user, ok := session.Values["user"].(model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	station, err := s.static.GetStationByID(user.StationID)
	if err != nil {
		internalServerError(w, "GetStationByID", err)
		fmt.Fprintln(w, "{}")
		return
	}

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["typeID"])

	isFavorite, err := s.db.IsFavorite(user.ID, id)
	if err != nil && err != model.ErrNotFound {
		internalServerError(w, "GetType", err)
		fmt.Fprintln(w, "{}")
		return
	}

	price, err := s.esi.MarketPrices(r.Context(), station.ID, station.Region.ID, id)
	if err != nil {
		internalServerError(w, "JitaPrices", err)
		fmt.Fprintln(w, "{}")
		return
	}

	history, err := s.esi.MarketHistory(r.Context(), station.Region.ID, id)
	if err != nil {
		internalServerError(w, "JitaHistory", err)
		fmt.Fprintln(w, "{}")
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

func (s *Server) TypeSetFavorite(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		fmt.Fprintln(w, "{}")
		return
	}

	user, ok := session.Values["user"].(model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	var req struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "{}")
		return
	}

	if err := s.db.SetFavorite(user.ID, typeID, req.Favorite); err != nil {
		internalServerError(w, "SetFavorite", err)
		fmt.Fprintln(w, "{}")
		return
	}

	json.NewEncoder(w).Encode(&req)
}

func (s *Server) TypeGetFavorites(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		fmt.Fprintln(w, "{}")
		return
	}

	user, ok := session.Values["user"].(model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	types, err := s.db.FavoriteTypes(user.ID)
	if err != nil {
		internalServerError(w, "FavoriteTypes", err)
		fmt.Fprintln(w, "{}")
		return
	}

	json.NewEncoder(w).Encode(&types)
}

func (s *Server) TypeOpenInGame(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		internalServerError(w, "store.Get", err)
		fmt.Fprintln(w, "{}")
		return
	}

	oldTok, ok := session.Values["token"].(oauth2.Token)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	tokSrc := oauthConfig.TokenSource(r.Context(), &oldTok)
	newTok, err := tokSrc.Token()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
		return
	}

	ctx := context.WithValue(r.Context(), ESITokenKey, newTok.AccessToken)

	vars := mux.Vars(r)
	typeID, _ := strconv.Atoi(vars["typeID"])

	if err := s.esi.OpenMarketWindow(ctx, typeID); err != nil {
		internalServerError(w, "OpenMarketWindow", err)
	} else {
		w.Header()["Content-Type"] = nil
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) TypeSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.static.SearchTypesByName(vars["filter"])
	if err != nil {
		internalServerError(w, "GetMarketTypes", err)
		fmt.Fprintln(w, "{}")
		return
	}

	json.NewEncoder(w).Encode(items)
}

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
