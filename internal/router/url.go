package router

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-http-utils/headers"
)

const urlScheme = "http://"

func (router *Router) getURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	originalURL, err := router.urlHandler.GetOriginal(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set(headers.Location, originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (router *Router) addURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	host := urlScheme + r.Host

	shortURL, err := router.urlHandler.Shorten(host, string(url))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
