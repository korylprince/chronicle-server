package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

type forwardedHandler struct {
	chain http.Handler
}

// ForwardedHandler replaces the Remote Address with the X-Forwarded-For header if it exists
func ForwardedHandler(h http.Handler) http.Handler {
	return forwardedHandler{h}
}

func (h forwardedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	_, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Panicln("Error parsing Remote Address:", err)
	}

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		r.RemoteAddr = fmt.Sprintf("%s:%s", ip, port)
	}

	h.chain.ServeHTTP(rw, r)
}

// StatsHandler returns the current stats
func StatsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	err := e.Encode(httpstats.Data())
	if err != nil {
		log.Println("Error encoding data:", err)
	}
}
