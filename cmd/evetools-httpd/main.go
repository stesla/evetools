package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"golang.org/x/oauth2"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
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

var cfgFile string

func init() {
	pflag.StringVar(&cfgFile, "config", "", "config file, default: $HOME/.evetools.yaml")

	pflag.String("addr", ":8080", "address to listen on")
	pflag.String("dir", "./public", "directory to serve files from")

	viper.BindPFlag("httpd.address", pflag.Lookup("addr"))
	viper.BindPFlag("httpd.static.dir", pflag.Lookup("dir"))
	viper.SetDefault("httpd.session.auth_key", securecookie.GenerateRandomKey(64))
	viper.SetDefault("httpd.session.name", "evetools")
	viper.SetDefault("model.database", "./evetools.sqlite3")
	viper.SetDefault("esi.basePath", "https://esi.evetech.net")
	viper.SetDefault("oauth.basePath", "https://login.eveonline.com")
	viper.SetDefault("http.cache.dir", "./cache")
	viper.SetDefault("sde.dir", "./data")
	viper.SetDefault("template.dir", "./public/templates")

	gob.Register(oauth2.Token{})
	gob.Register(model.User{})
}

var oauthConfig = oauth2.Config{
	Scopes: []string{
		"esi-markets.read_character_orders.v1",
		"esi-ui.open_window.v1",
		"esi-characters.read_standings.v1",
		"esi-skills.read_skills.v1",
		"esi-wallet.read_character_wallet.v1",
		"publicData",
	},
}

func main() {
	pflag.Parse()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".evetools")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("error loading config file: %s", err)
	}

	db, err := model.Initialize(viper.GetString("model.database"))
	if err != nil {
		log.Fatalf("error initializing model: %s", err)
	}

	if err := initOAuthConfig(); err != nil {
		log.Fatalf("error initializing oauth: %s", err)
	}

	store = sessions.NewCookieStore([]byte(viper.GetString("httpd.session.auth_key")))

	staticFiles := http.Dir(viper.GetString("httpd.static.dir"))

	views := &templateViewRenderer{viper.GetString("template.dir")}

	var handler http.Handler = NewServer(http.FileServer(staticFiles), db, views)
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
	http http.Client
	esi  esi.Client
	mux  *mux.Router
	db   model.DB

	viewRenderer
}

func NewServer(static http.Handler, db model.DB, vr viewRenderer) *Server {
	s := &Server{
		mux:          mux.NewRouter(),
		db:           db,
		viewRenderer: vr,
	}
	s.esi = esi.NewClient(&s.http)

	cache := diskcache.New(viper.GetString("http.cache.dir"))
	s.http.Transport = httpcache.NewTransport(cache)

	s.mux.NotFoundHandler = http.HandlerFunc(s.NotFound)
	s.mux.Use(s.haveSession)
	s.mux.Use(s.haveLoggedInUser)

	// Static
	s.mux.PathPrefix("/css").Handler(static)
	s.mux.PathPrefix("/data").Handler(static)
	s.mux.PathPrefix("/js").Handler(static)
	s.mux.PathPrefix("/views").Handler(static)

	// Login Stuff
	s.mux.Methods("GET").Path("/login").HandlerFunc(s.Login)
	s.mux.Methods("GET").Path("/login/authorize").HandlerFunc(s.Authorize)
	s.mux.Methods("GET").Path("/login/callback").HandlerFunc(s.LoginCallback)
	s.mux.Methods("GET").Path("/logout").HandlerFunc(s.Logout)

	// Views
	s.mux.Methods("GET").Path("/").Handler(s.ShowView("dashboard"))
	s.mux.Methods("GET").Path("/authorize").Handler(s.ShowView("authorize"))
	s.mux.Methods("GET").Path("/browse").HandlerFunc(s.ShowBrowse)
	s.mux.Methods("GET").Path("/groups/{groupID:[0-9]+}").HandlerFunc(s.ShowGroupDetails)
	s.mux.Methods("GET").Path("/history").Handler(s.ShowView("orders"))
	s.mux.Methods("GET").Path("/orders").Handler(s.ShowView("orders"))
	s.mux.Methods("GET").Path("/search").HandlerFunc(s.ShowSearch)
	s.mux.Methods("GET").Path("/settings").Handler(s.ShowView("settings"))
	s.mux.Methods("GET").Path("/transactions").Handler(s.ShowView("transactions"))
	s.mux.Methods("GET").Path("/types/{typeID:[0-9]+}").Handler(s.ShowView("typeDetails"))

	// API
	api := s.mux.PathPrefix("/api/v1").Subrouter()
	api.Use(contentType("application/json").Middleware)
	api.Methods("GET").Path("/stations").HandlerFunc(s.GetStations)
	api.Methods("PUT").Path("/types/{typeID:[0-9]+}/favorite").HandlerFunc(s.PutTypeFavorite)
	api.Methods("POST").Path("/types/{typeID:[0-9]+}/openInGame").HandlerFunc(s.PostOpenInGame)
	api.Methods("DELETE").Path("/user/characters/{characterID:[0-9]+}").
		HandlerFunc(s.DeleteUserCharacter)
	api.Methods("POST").Path("/user/characters/{characterID:[0-9]+}/activate").
		HandlerFunc(s.PostUserCharacterActivate)
	api.Methods("GET").Path("/user/favorites").HandlerFunc(s.GetUserFavorites)
	api.Methods("PUT").Path("/user/stationA").HandlerFunc(s.PutUserStationA)
	api.Methods("PUT").Path("/user/stationB").HandlerFunc(s.PutUserStationB)

	// View Data
	view := api.PathPrefix("/view").Subrouter()
	view.Methods("GET").Path("/dashboard").HandlerFunc(s.ViewDashboard)
	view.Methods("GET").Path("/marketOrders").HandlerFunc(s.ViewMarketOrders)
	view.Methods("GET").Path("/settings").HandlerFunc(s.ViewSettings)
	view.Methods("GET").Path("/transactions").HandlerFunc(s.ViewTransactions)
	view.Methods("GET").Path("/typeDetails/{typeID:[0-9]+}").HandlerFunc(s.ViewTypeDetails)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, &s.http)
	s.mux.ServeHTTP(w, r.WithContext(ctx))
}

func (s *Server) Authorize(w http.ResponseWriter, r *http.Request) {
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
	var scopelessConfig = oauthConfig
	scopelessConfig.Scopes = nil
	url := scopelessConfig.AuthCodeURL(state)
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
	jwt, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		internalServerError(w, "oauth.Exchange", err)
		return
	}

	ctx := context.WithValue(r.Context(), esi.AccessTokenKey, jwt.AccessToken)
	verify, err := s.esi.Verify(ctx)
	if err != nil {
		internalServerError(w, "Verify", err)
		return
	}

	var user *model.User
	var character *model.Character

	if sessionUserID, ok := session.Values["userid"].(int); ok {
		user, err = s.db.GetUser(sessionUserID)
		if err != nil {
			internalServerError(w, "Getuser", err)
			return
		}
		character, err = s.db.FindOrCreateCharacterForUser(user.ID, verify)
		if err != nil {
			internalServerError(w, "FindOrCreateCharacterForUser", err)
			return
		}
	} else {
		user, character, err = s.db.FindOrCreateUserAndCharacter(verify)
		if err != nil {
			internalServerError(w, "FindOrCreateUserAndCharacter", err)
			return
		}
	}

	if verify.Scopes != "" {
		err := s.db.SaveTokenForCharacter(character.ID, verify.Scopes, jwt.RefreshToken)
		if err != nil {
			internalServerError(w, "SaveTokenForCharacter", err)
			return
		}
	}

	var needAuth bool = false

	character, err = s.db.GetCharacterByOwnerHash(user.ActiveCharacterHash)
	if err != nil {
		internalServerError(w, "GetCharacterByOwnerHash", err)
		return
	}

	token, err := s.db.GetTokenForCharacter(character.ID)
	if err == model.ErrNotFound {
		needAuth = true
	} else if err != nil {
		internalServerError(w, "GetTokenForCharacter", err)
		return
	} else if hasScopes(token.Scopes, oauthConfig.Scopes) {
		jwt, err = refreshToken(r.Context(), token.RefreshToken)
		if err != nil {
			internalServerError(w, "refreshToken", err)
			return
		}
		err := s.db.SaveTokenForCharacter(character.ID, token.Scopes, jwt.RefreshToken)
		if err != nil {
			internalServerError(w, "SaveTokenForCharacter", err)
			return
		}
	} else {
		needAuth = true
	}

	session.Values["token"] = jwt
	session.Values["userid"] = user.ID
	if err := session.Save(r, w); err != nil {
		internalServerError(w, "save session", err)
		return
	}

	var next string
	if needAuth {
		next = "/authorize"
	} else {
		next = "/"
		if str, ok := session.Values["next"].(string); ok && str != "" {
			next = str
		}
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func hasScopes(a string, bs []string) bool {
	as := strings.Fields(a)
	if len(as) != len(bs) {
		return false
	}
	sort.Strings(as)
	sort.Strings(bs)
	for i := 0; i < len(bs); i++ {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	session := currentSession(r)
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) NotFound(w http.ResponseWriter, r *http.Request) {
	s.renderView(w, r, "notFound", nil, nil)
}

func (s *Server) ShowView(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.renderView(w, r, name, nil, nil)
	})
}

type viewRenderer interface {
	renderView(w http.ResponseWriter, r *http.Request, name string, helpers template.FuncMap, data interface{})
}

type templateViewRenderer struct {
	dir string
}

func (tvr *templateViewRenderer) renderView(w http.ResponseWriter, r *http.Request, name string, helpers template.FuncMap, data interface{}) {
	funcs := template.FuncMap{
		"avatarURL": func(id int) string {
			return fmt.Sprintf("https://images.evetech.net/characters/%d/portrait?size=128", id)
		},
		"currentUser": func() (u *model.User) {
			defer func() {
				if r := recover(); r != nil {
					u = nil
				}
			}()
			u = currentUser(r)
			return
		},
		"iconURL": func(t *sde.MarketType) string {
			if t == nil {
				return ""
			}
			imgType := "icon"
			if strings.Contains(t.Name, "Blueprint") || strings.Contains(t.Name, "Formula") {
				imgType = "bp"
			}
			return fmt.Sprintf("https://images.evetech.net/types/%d/%s?size=128", t.ID, imgType)
		},
	}
	for k, v := range helpers {
		funcs[k] = v
	}
	t := template.New("template").Funcs(funcs)
	tpl, err := tvr.loadTemplate("layout.html")
	if err != nil {
		internalServerError(w, "loadTemplate", err)
		return
	}
	t = template.Must(t.Parse(tpl))
	tpl, err = tvr.loadTemplate(name + ".html")
	if err != nil {
		internalServerError(w, "loadTemplate", err)
		return
	}
	t = template.Must(t.Parse(tpl))
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		// No internalServerError here because if we wrote to w in
		// ExecuteTemplate, it's already set the status code to 200, an setting
		// it again is an error.
		log.Println("ExecuteTemplate:", err)
	}
}

func (t *templateViewRenderer) loadTemplate(filename string) (string, error) {
	filepath := path.Join(t.dir, filename)
	input, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer input.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(input)
	return buf.String(), err
}
