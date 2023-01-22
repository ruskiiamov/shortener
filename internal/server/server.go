package server

import (
	"net/http"

	"github.com/ruskiiamov/shortener/internal/url"
)

type Router interface {
	http.Handler
	GET(pattern string, handler http.Handler)
	POST(pattern string, handler http.Handler)
	DELETE(pattern string, handler http.Handler)
	GetURLParam(r *http.Request, key string) string
	AddMiddlewares(middlewares ...func(http.Handler) http.Handler)
}

type Config struct {
	BaseURL string
	SignKey string
}

type handler struct {
	router       Router
	urlConverter url.Converter
	baseURL      string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func NewHandler(u url.Converter, r Router, c Config) *handler {
	h := &handler{
		router:       r,
		urlConverter: u,
		baseURL:      c.BaseURL,
	}

	initAuth(c.SignKey)
	h.router.AddMiddlewares(gzipCompress, auth)

	h.router.GET("/{id}", h.getURL())
	h.router.POST("/", h.addURL())
	h.router.POST("/api/shorten", h.addURLFromJSON())
	h.router.POST("/api/shorten/batch", h.addURLBatch())
	h.router.GET("/api/user/urls", h.getAllURL())
	h.router.DELETE("/api/user/urls", h.deleteURLBatch())
	h.router.GET("/ping", h.pingDB())

	return h
}
