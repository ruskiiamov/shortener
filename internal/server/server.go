// Server is the handler mux for all HTTP requests.
package server

import (
	"context"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/url"
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

// Server config contains base URL and sign key for authorization.
type Config struct {
	// Base server URL
	BaseURL string

	// Key for HMAC sign
	SignKey string
}

type handler struct {
	router       Router
	urlConverter url.Converter
	baseURL      string
	delBuf       chan *delBatch
	delFinish    chan struct{}
}

// ServeHTTP is the method of the http.Handler interface.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// Close closes all server goroutines.
func (h *handler) Close(ctx context.Context) error {
	close(h.delBuf)

	for {
		select {
		case <-h.delFinish:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// NewHandler returns handler mux for HTTP server
func NewHandler(ctx context.Context, u url.Converter, r Router, c Config) *handler {
	h := &handler{
		router:       r,
		urlConverter: u,
		baseURL:      c.BaseURL,
		delBuf:       make(chan *delBatch),
		delFinish:    make(chan struct{}),
	}

	go h.startDeleteURL(ctx)

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
