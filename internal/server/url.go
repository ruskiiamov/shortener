package server

import (
	"io"
	"net/http"

	"github.com/go-http-utils/headers"
)

const urlScheme = "http://"

func (h *Handler) getURL(w http.ResponseWriter, r *http.Request) {
	id := h.router.GetURLParam(r, "id")

	originalURL, err := h.converter.GetOriginal(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set(headers.Location, originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) addURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	host := urlScheme + r.Host

	shortURL, err := h.converter.Shorten(host, string(url))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
