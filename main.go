package main

//go:generate go-bindata-assetfs static/...

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/korylprince/chronicle-server/api"
	"github.com/thoas/stats"
)

var httpstats *stats.Stats

func init() {
	httpstats = stats.New()
}

//middleware
func middleware(h http.Handler) http.Handler {
	return httpstats.Handler(handlers.CombinedLoggingHandler(os.Stdout,
		handlers.CompressHandler(
			handlers.CORS(
				handlers.AllowedOrigins([]string{"*"}),
				handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
				handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Origin"}),
			)(
				http.StripPrefix(config.Prefix,
					ForwardedHandler(h))))))
}

func main() {
	db, err := api.NewDB(config.SQLDriver, config.SQLDSN, config.Workers, time.Duration(config.WriteInterval)*time.Second)
	if err != nil {
		log.Panicln("Error creating DB:", err)
	}

	c := &api.Context{
		DB: db,
	}

	r := mux.NewRouter()

	//api
	r.Handle("/api/v1/submit", api.SubmitHandler(c)).Methods("POST")
	r.Handle("/api/v1.1/submit", api.SubmitHandler(c)).Methods("POST")
	r.Handle("/api/v1.1/stats", http.HandlerFunc(StatsHandler)).Methods("GET")

	log.Println("Listening on:", config.ListenAddr)
	log.Println(http.ListenAndServe(config.ListenAddr, middleware(r)))
}
