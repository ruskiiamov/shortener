// Package router is a HTTP router
package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type router struct {
	mux *chi.Mux
}

// NewRouter returns a new router object that implements http.Handler interface.
func NewRouter() *router {
	chiMux := chi.NewMux()

	chiMux.Use(middleware.Logger, middleware.Recoverer)

	return &router{mux: chiMux}
}

// ServeHTTP is the method of the http.Handler interface.
func (rtr *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr.mux.ServeHTTP(w, r)
}

// GET registers hanlders for GET HTTP method.
func (rtr *router) GET(pattern string, handler http.Handler) {
	rtr.mux.Get(pattern, handler.ServeHTTP)
}

// POST registers hanlders for POST HTTP method.
func (rtr *router) POST(pattern string, handler http.Handler) {
	rtr.mux.Post(pattern, handler.ServeHTTP)
}

// DELETE registers hanlders for DELETE HTTP method.
func (rtr *router) DELETE(pattern string, handler http.Handler) {
	rtr.mux.Delete(pattern, handler.ServeHTTP)
}

// GetURLParam returns the URL parameter.
func (rtr *router) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// AddMiddlewares regiters global middlewares into the chain.
func (rtr *router) AddMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	rtr.mux.Use(middlewares...)
}
