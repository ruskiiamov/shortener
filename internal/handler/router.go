package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Shortener interface {
	Shorten(host, url string) (string, error)
	GetOriginal(id string) (string, error)
}

type Handler struct {
	*chi.Mux
	shortener Shortener
}

func New(s Shortener) *Handler {
	h := &Handler{
		Mux:       chi.NewMux(),
		shortener: s,
	}

	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)

	h.Get("/{id}", h.getURL())
	h.Post("/", h.addURL())

	return h
}
