package server

import "net/http"

type Router interface {
	http.Handler
	GET(pattern string, handlerFn http.HandlerFunc)
	POST(pattern string, handlerFn http.HandlerFunc)
	GetURLParam(r *http.Request, key string) string
}

type Converter interface {
	Shorten(url string) (string, error)
	GetOriginal(id string) (string, error)
}

type Handler struct {
	router    Router
	converter Converter
}

func NewHandler(c Converter, r Router) *Handler {
	h := &Handler{
		router:    r,
		converter: c,
	}

	h.router.GET("/{id}", h.getURL)
	h.router.POST("/", h.addURL)
	h.router.POST("/api/shorten", h.addURLFromJSON)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
