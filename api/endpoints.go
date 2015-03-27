package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"
)

//submitHandler takes an Entry and commits it to the DB
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
