package api

import "net/http"

// Context is a set of services accessed by endpoints
type Context struct {
	DB     *DB
	APIKey string
}

type contextHandler struct {
	HandleFunc func(*Context, http.ResponseWriter, *http.Request)
	Context    *Context
}

func (c contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.HandleFunc(c.Context, w, r)
}

// SubmitHandler returns an http.Handler for submitHandler
func SubmitHandler(c *Context) http.Handler {
	return contextHandler{HandleFunc: submitHandler, Context: c}
}
