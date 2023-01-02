package server

import (
	"net/http"

	"github.com/ruskiiamov/shortener/internal/url"
)

type Router interface {
	http.Handler
	GET(pattern string, handler http.Handler)
	POST(pattern string, handler http.Handler)
	GetURLParam(r *http.Request, key string) string
	AddMiddlewares(middlewares ...func(http.Handler) http.Handler)
}

type handler struct {
	router       Router
	urlConverter url.Converter
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func NewHandler(u url.Converter, r Router, signKey string) *handler {
	h := &handler{
		router:       r,
		urlConverter: u,
	}

	initAuth(signKey)
	h.router.AddMiddlewares(gzipCompress, auth)

	h.router.GET("/{id}", h.getURL())
	h.router.POST("/", h.addURL())
	h.router.POST("/api/shorten", h.addURLFromJSON())
	h.router.GET("/api/user/urls", h.getAllURL())

	return h
}
