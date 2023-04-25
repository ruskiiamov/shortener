// Package server is the handler mux for all HTTP requests.
package server

import (
	"context"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/ruskiiamov/shortener/internal/user"
)

// Router is used by server to set all handlers and middlewares.
type Router interface {
	http.Handler
	GET(pattern string, handler http.Handler)
	POST(pattern string, handler http.Handler)
	DELETE(pattern string, handler http.Handler)
	GetURLParam(r *http.Request, key string) string
	AddMiddlewares(middlewares ...func(http.Handler) http.Handler)
}

type handler struct {
	router       Router
	urlConverter url.Converter
	baseURL      string
	delBuf       chan *url.DelBatch
}

// ServeHTTP is the method of the http.Handler interface.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// NewHandler returns handler mux for HTTP server
func NewHandler(ctx context.Context, ua user.Authorizer, uc url.Converter, r Router, delBuf chan *url.DelBatch, baseURL, cidr string) (*handler, error) {
	h := &handler{
		router:       r,
		urlConverter: uc,
		baseURL:      baseURL,
		delBuf:       delBuf,
	}

	err := setCIDR(cidr)
	if err != nil {
		return nil, err
	}

	h.router.AddMiddlewares(
		trustedSubnet,
		gzipCompress,
		newAuthMiddleware(ua).handle,
	)

	h.router.GET("/{id}", h.getURL())
	h.router.POST("/", h.addURL())
	h.router.POST("/api/shorten", h.addURLFromJSON())
	h.router.POST("/api/shorten/batch", h.addURLBatch())
	h.router.GET("/api/user/urls", h.getAllURL())
	h.router.DELETE("/api/user/urls", h.deleteURLBatch())
	h.router.GET("/api/internal/stats", h.stats())
	h.router.GET("/ping", h.pingDB())

	return h, nil
}
