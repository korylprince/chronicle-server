package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// submitHandler takes an Entry and commits it to the DB
func submitHandler(c *Context, rw http.ResponseWriter, r *http.Request) {
	e := &Entry{}
	d := json.NewDecoder(r.Body)
	err := d.Decode(e)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Panicln("Error parsing Remote Address:", err)
	}

	e.InternetIP = ip
	e.Time = time.Now()

	err = e.Validate()
	if err != nil {
		log.Println("Validation Error:", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	c.DB.Push(e)
}

// API errors
var (
	ErrAPINotEnabled      = errors.New("api not enabled")
	ErrInvalidAuthHeader  = errors.New("invalid auth header")
	ErrInvalidAPIKey      = errors.New("invalid api key")
	ErrInvalidSerialCount = errors.New("invalid serial count")
)

func (c *Context) handleQueryLastUser(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if c.APIKey == "" {
		return http.StatusNotFound, ErrAPINotEnabled
	}
	header := strings.Split(r.Header.Get("Authorization"), " ")
	if len(header) != 2 || header[0] != "Bearer" || len(header[1]) == 0 {
		return http.StatusUnauthorized, ErrInvalidAuthHeader
	}

	if subtle.ConstantTimeEq(int32(len(c.APIKey)), int32(len([]byte(header[1])))) != 1 ||
		subtle.ConstantTimeCompare([]byte(c.APIKey), []byte(header[1])) != 1 {
		return http.StatusUnauthorized, ErrInvalidAPIKey
	}

	var serials []string
	if err := json.NewDecoder(r.Body).Decode(&serials); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not parse body: %w", err)
	}
	if len(serials) == 0 {
		return http.StatusBadRequest, ErrInvalidSerialCount
	}

	entries, err := c.DB.QueryLastUser(serials)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("could not query database: %w", err)
	}

	return http.StatusOK, entries
}

// HandleQueryLastUser returns the latest information for the submitted serials
func (c *Context) HandleQueryLastUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, body := c.handleQueryLastUser(w, r)
		w.WriteHeader(status)

		if err, ok := body.(error); ok {
			log.Println("api error:", err)
			body = map[string]interface{}{"status": status}
		}

		if err := json.NewEncoder(w).Encode(body); err != nil {
			log.Println("couldn't encode body:", err)
		}
	})
}
