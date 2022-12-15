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

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
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

	h.router.GET("/{id}", Conveyor(h.getURL, gzipCompress))
	h.router.POST("/", Conveyor(h.addURL, gzipCompress))
	h.router.POST("/api/shorten", Conveyor(h.addURLFromJSON, gzipCompress))

	return h
}
