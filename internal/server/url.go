package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-http-utils/headers"
	"github.com/ruskiiamov/shortener/internal/url"
)

const applicationJSON = "application/json"

type requestData struct {
	URL string `json:"url"`
}

type responseData struct {
	Result string `json:"result"`
}

type requestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type responseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type responseAll struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *handler) getURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := h.router.GetURLParam(r, "id")

		shortURL, err := h.urlConverter.GetOriginal(id)
		if errors.Is(err, new(url.ErrURLDeleted)) {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add(headers.Location, shortURL.Original)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

func (h *handler) addURL() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var errDupl *url.ErrURLDuplicate

		shortURL, err := h.urlConverter.Shorten(userID.Value, string(body))
		if errors.As(err, &errDupl) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(h.baseURL + "/" + errDupl.EncodedID))
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(h.baseURL + "/" + shortURL.EncodedID))
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

		var errDupl *url.ErrURLDuplicate

		shortURL, err := h.urlConverter.Shorten(userID.Value, reqData.URL)
		if errors.As(err, &errDupl) {
			resData := responseData{h.baseURL + "/" + errDupl.EncodedID}
			jsonRes, err := json.Marshal(resData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Add(headers.ContentType, applicationJSON)
			w.WriteHeader(http.StatusConflict)
			w.Write(jsonRes)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resData := responseData{h.baseURL + "/" + shortURL.EncodedID}
		jsonRes, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add(headers.ContentType, applicationJSON)
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonRes)
	})
}

func (h *handler) addURLBatch() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var reqData []requestBatch
		if err := json.Unmarshal(body, &reqData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var originals []string
		for _, item := range reqData {
			originals = append(originals, item.OriginalURL)
		}
		shortURLs, err := h.urlConverter.ShortenBatch(userID.Value, originals)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var resData []responseBatch
		for _, item := range reqData {
			for _, shortURL := range shortURLs {
				if shortURL.Original == item.OriginalURL {
					resData = append(resData, responseBatch{
						CorrelationID: item.CorrelationID,
						ShortURL:      h.baseURL + "/" + shortURL.EncodedID,
					})
					break
				}
			}
		}

		if len(reqData) != len(resData) {
			http.Error(w, "url adding error", http.StatusInternalServerError)
			return
		}

		jsonRes, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		shortURLs, err := h.urlConverter.GetAllByUser(userID.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(shortURLs) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var resData []responseAll
		for _, shortURL := range shortURLs {
			resData = append(resData, responseAll{
				ShortURL:    h.baseURL + "/" + shortURL.EncodedID,
				OriginalURL: shortURL.Original,
			})
		}

		jsonRes, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add(headers.ContentType, applicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonRes)
	})
}

func (h *handler) deleteURLBatch() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var encodedIDs []string
		if err := json.Unmarshal(body, &encodedIDs); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := r.Cookie(userIDCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go func() {
			err = h.urlConverter.RemoveBatch(userID.Value, encodedIDs)
			if err != nil {
				log.Printf("delete URL batch error: userID=%v encodedIDs=%v", userID.Value, encodedIDs)
				return
			}
			log.Printf("URL batch deleted: userID=%v encodedIDs=%v", userID.Value, encodedIDs)
		}()

		w.WriteHeader(http.StatusAccepted)
	})
}

func (h *handler) pingDB() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h.urlConverter.PingKeeper(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
