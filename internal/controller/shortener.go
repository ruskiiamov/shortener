package controller

import (
	"io"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/usecase"
)

func newShortenerHandler(s usecase.Shortener) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			get(s, w, r)
		case http.MethodPost:
			post(s, w, r)
		default:
			http.Error(w, "wrong http method", http.StatusBadRequest)
		}
	}
}

func get(s usecase.Shortener, w http.ResponseWriter, r *http.Request) {
	var id string

	uri := r.RequestURI
	if uri[0] == '/' {
		id = uri[1:]
	} else {
		id = uri
	}
	
	originalURL, err := s.GetOriginal(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func post(s usecase.Shortener, w http.ResponseWriter, r *http.Request) {
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	shortenedURL, err := s.Shorten(string(originalURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://" + r.Host + "/" + shortenedURL))
}