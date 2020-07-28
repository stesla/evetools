package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	addr      = flag.String("addr", ":8080", "address to listen on")
	staticDir = flag.String("dir", "./public", "directory to serve static files from")
)

func main() {
	flag.Parse()

	router := mux.NewRouter()
	router.Methods("GET").Path("/api/v1/ping").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintln(w, `["pong"]`)
	})
	// this needs to be last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(*staticDir)))

	var handler http.Handler = router
	handler = handlers.LoggingHandler(os.Stdout, handler)
	handler = handlers.ProxyHeaders(handler)

	log.Println("listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
