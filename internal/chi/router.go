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
	chiMux.Use(middleware.Recoverer)

	return &router{mux: chiMux}
}

func (rtr *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rtr.mux.ServeHTTP(w, r)
}

func (rtr *router) GET(pattern string, handlerFn http.HandlerFunc) {
	rtr.mux.Get(pattern, handlerFn)
}

func (rtr *router) POST(pattern string, handlerFn http.HandlerFunc) {
	rtr.mux.Post(pattern, handlerFn)
}

func (rtr *router) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
