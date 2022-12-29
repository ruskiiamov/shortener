package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type router struct {
	mux *chi.Mux
}

func NewRouter() *router {
	chiMux := chi.NewMux()

	chiMux.Use(middleware.Logger)

	return &router{mux: chiMux}
}

func (rtr *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr.mux.ServeHTTP(w, r)
}

func (rtr *router) GET(pattern string, handler http.Handler) {
	rtr.mux.Get(pattern, handler.ServeHTTP)
}

func (rtr *router) POST(pattern string, handler http.Handler) {
	rtr.mux.Post(pattern, handler.ServeHTTP)
}

func (rtr *router) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func (rtr *router) AddMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	rtr.mux.Use(middlewares...)
}
