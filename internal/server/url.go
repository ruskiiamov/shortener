package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-http-utils/headers"
)

const applicationJSON = "application/json"

type url struct {
	URL string `json:"url"`
}

type result struct {
	Result string `json:"result"`
}

func (h *Handler) getURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := h.router.GetURLParam(r, "id")

		originalURL, err := h.converter.GetOriginal(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add(headers.Location, originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

func (h *Handler) addURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortURL, err := h.converter.Shorten(string(url))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	})
}

func (h *Handler) addURLFromJSON() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u := new(url)
		if err := json.Unmarshal(body, u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortURL, err := h.converter.Shorten(string(u.URL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res := result{shortURL}
		jsonRes, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add(headers.ContentType, applicationJSON)
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonRes)
	})
}
