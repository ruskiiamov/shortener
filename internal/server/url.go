package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-http-utils/headers"
)

const applicationJSON = "application/json"

type requestData struct {
	URL string `json:"url"`
}

type responseData struct {
	Result string `json:"result"`
}

func (h *handler) getURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := h.router.GetURLParam(r, "id")

		originalURL, err := h.urlConverter.GetOriginal(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add(headers.Location, originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

func (h *handler) addURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL, err := h.urlConverter.Shorten(string(url), userID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	})
}

func (h *handler) addURLFromJSON() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reqData := new(requestData)
		if err := json.Unmarshal(body, reqData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL, err := h.urlConverter.Shorten(reqData.URL, userID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resData := responseData{shortURL}
		jsonRes, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add(headers.ContentType, applicationJSON)
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonRes)
	})
}

func (h *handler) getAllURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urls, err := h.urlConverter.GetAll(userID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		jsonRes, err := json.Marshal(urls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add(headers.ContentType, applicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

func (h *handler) pingDB() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h.urlConverter.PingDB(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
