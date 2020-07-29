package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	addr      = flag.String("addr", ":8080", "address to listen on")
	staticDir = flag.String("dir", "./public", "directory to serve static files from")
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var sessionName = getenvDefault("SESSION_NAME", "evetools-session")

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.Methods("GET").Path("/api/v1/currentUser").HandlerFunc(CurrentUser)
	// this needs to be last
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(*staticDir)))

	var handler http.Handler = makeHandler(http.FileServer(http.Dir(*staticDir)))
	handler = handlers.LoggingHandler(os.Stdout, handler)
	handler = handlers.ProxyHeaders(handler)

	log.Println("listening on", *addr)

	log.Fatal(http.ListenAndServe(*addr, handler))
}

func makeHandler(static http.Handler) http.Handler {
	r := mux.NewRouter()
	r.Methods("GET").Path("/api/v1/currentUser").HandlerFunc(CurrentUser)
	// this needs to be last
	r.PathPrefix("/").Handler(static)
	return r
}

func CurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	log.Println(r.Cookies())

	session, err := store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	user, ok := session.Values["currentUser"]
	if ok {
		json.NewEncoder(w).Encode(&user)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "{}")
	}
}

func getenvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	} else {
		return defaultVal
	}
}
