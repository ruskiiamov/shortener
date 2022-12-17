package server

import "net/http"

type Router interface {
	http.Handler
	GET(pattern string, handler http.Handler)
	POST(pattern string, handler http.Handler)
	GetURLParam(r *http.Request, key string) string
	AddMiddlewares(middlewares ...func(http.Handler) http.Handler)
}

type Converter interface {
	Shorten(url string) (string, error)
	GetOriginal(id string) (string, error)
}

type Handler struct {
	router    Router
	converter Converter
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func NewHandler(c Converter, r Router) *Handler {
	h := &Handler{
		router:    r,
		converter: c,
	}

	h.router.AddMiddlewares(gzipCompress)

	h.router.GET("/{id}", h.getURL())
	h.router.POST("/", h.addURL())
	h.router.POST("/api/shorten", h.addURLFromJSON())

	return h
}
