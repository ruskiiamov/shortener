package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URLHandler interface {
	Shorten(host, url string) (string, error)
	GetOriginal(id string) (string, error)
}

type Router struct {
	mux        *chi.Mux
	urlHandler URLHandler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func NewRouter(h URLHandler) *Router {
	r := &Router{
		mux:        chi.NewMux(),
		urlHandler: h,
	}

	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)

	r.mux.Get("/{id}", r.getURL)
	r.mux.Post("/", r.addURL)

	return r
}
