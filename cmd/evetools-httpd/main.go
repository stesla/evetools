package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/antihax/goesi"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var store *sessions.CookieStore

func init() {
	pflag.String("addr", ":8080", "address to listen on")
	pflag.String("dir", "./public", "directory to serve files from")

	viper.BindPFlag("httpd.address", pflag.Lookup("addr"))
	viper.BindPFlag("httpd.static.dir", pflag.Lookup("dir"))
	viper.SetDefault("httpd.session.auth_key", securecookie.GenerateRandomKey(64))
	viper.SetDefault("httpd.session.name", "evetools")

	gob.Register(oauth2.Token{})
	gob.Register(user{})
}

var eveclient *goesi.APIClient
var evesso *goesi.SSOAuthenticator
var scopes = []string{
	"publicData",
}

type user struct {
	CharacterID   int32  `json:"characterID"`
	CharacterName string `json:"characterName"`
}

func main() {
	pflag.Parse()

	viper.SetConfigName(".evetools")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Fatal error loading config file: %s", err)
	}

	clientID := viper.GetString("oauth.client_id")
	if clientID == "" {
		log.Fatalln("must provide oauth.client_id")
	}
	clientSecret := viper.GetString("oauth.client_secret")
	if clientSecret == "" {
		log.Fatalln("must provide oauth.client_secret")
	}
	redirectURL := viper.GetString("oauth.redirect_url")
	if redirectURL == "" {
		log.Fatalln("must provide oauth.redirect_url")
	}
	httpclient := &http.Client{}
	eveclient = goesi.NewAPIClient(httpclient, "evetools-dev (stesla@pobox.com - Stewart Cash)")
	evesso = goesi.NewSSOAuthenticatorV2(httpclient, clientID, clientSecret, redirectURL, scopes)

	store = sessions.NewCookieStore([]byte(viper.GetString("httpd.session.auth_key")))

	staticFiles := http.Dir(viper.GetString("httpd.static.dir"))
	var handler http.Handler = makeHandler(http.FileServer(staticFiles))
	handler = handlers.LoggingHandler(os.Stdout, handler)
	handler = handlers.ProxyHeaders(handler)

	addr := viper.GetString("httpd.address")
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func makeHandler(static http.Handler) http.Handler {
	r := mux.NewRouter()
	r.Methods("GET").Path("/login").HandlerFunc(Login)
	r.Methods("GET").Path("/login/callback").HandlerFunc(LoginCallback)
	r.Methods("GET").Path("/api/v1/currentUser").HandlerFunc(CurrentUser)
	// this needs to be last
	r.PathPrefix("/").Handler(static)
	return r
}

func CurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, ok := session.Values["currentUser"]
	if ok {
		json.NewEncoder(w).Encode(user)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, viper.GetString("httpd.session.name"))
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.RawURLEncoding.EncodeToString(b)
	session.Values["oauth2.state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := evesso.AuthorizeURL(state, true, scopes)
	http.Redirect(w, r, url, http.StatusFound)
}

func LoginCallback(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, viper.GetString("httpd.session.name"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := r.FormValue("state")
	if str, ok := session.Values["oauth2.state"].(string); !ok || str != state {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code := r.FormValue("code")
	token, err := evesso.TokenExchange(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokSrc := evesso.TokenSource(token)
	v, err := evesso.Verify(tokSrc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["currentUser"] = user{
		CharacterID:   v.CharacterID,
		CharacterName: v.CharacterName,
	}
	session.Values["token"] = token
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	next := "/"
	if str, ok := session.Values["next"].(string); ok && str != "" {
		next = str
	}
	http.Redirect(w, r, next, http.StatusFound)
}
