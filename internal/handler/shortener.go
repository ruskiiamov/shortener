package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) getURL() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		originalURL, err := h.shortener.GetOriginal(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) addURL() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		originalURL, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		host := "http://" + r.Host

		shortenedURL, err := h.shortener.Shorten(host, string(originalURL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortenedURL))
	}
}
