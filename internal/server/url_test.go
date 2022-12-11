package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/stretchr/testify/assert"
)

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		id      string
		res     string
		err     error
		wantErr bool
	}{
		{
			name:    "ok",
			path:    "/0",
			id:      "0",
			res:     "https://google.com",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "not ok",
			path:    "/abc",
			id:      "abc",
			res:     "",
			err:     errors.New("wrong id"),
			wantErr: true,
		},
	}

	mockedURLHandler := new(MockedURLHandler)
	h := NewHandler(mockedURLHandler, chi.NewRouter())
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedURLHandler.On("GetOriginal", tt.id).Return(tt.res, tt.err)

			statusCode, _, header := testRequest(t, ts, http.MethodGet, tt.path, nil)

			mockedURLHandler.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, statusCode)
				return
			}

			assert.Equal(t, http.StatusTemporaryRedirect, statusCode)
			assert.Equal(t, tt.res, header.Get("Location"))
		})
	}
}

func TestAddURL(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		res     string
		err     error
		wantErr bool
	}{
		{
			name:    "ok",
			body:    "http://shortener.com",
			res:     "/0",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "not ok",
			body:    "shortener.com",
			res:     "",
			err:     errors.New("wrong url"),
			wantErr: true,
		},
	}

	mockedURLHandler := new(MockedURLHandler)
	h := NewHandler(mockedURLHandler, chi.NewRouter())
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedURLHandler.On("Shorten", ts.URL, tt.body).Return(ts.URL+tt.res, tt.err)

			statusCode, body, _ := testRequest(t, ts, http.MethodPost, "/", []byte(tt.body))

			mockedURLHandler.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, statusCode)
				return
			}

			assert.Equal(t, http.StatusCreated, statusCode)
			assert.Equal(t, ts.URL+tt.res, string(body))
		})
	}
}

func TestAddURLFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		res     string
		cType   string
		err     error
		wantErr bool
	}{
		{
			name:    "ok",
			url:     "http://shortener.com",
			res:     "/0",
			cType:   "application/json",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "not ok",
			url:     "shortener.com",
			res:     "",
			cType:   "",
			err:     errors.New("wrong url"),
			wantErr: true,
		},
	}

	mockedURLHandler := new(MockedURLHandler)
	h := NewHandler(mockedURLHandler, chi.NewRouter())
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody := `{"url":"` + tt.url + `"}`
			jsonResp := `{"result":"` + ts.URL + tt.res + `"}`
			mockedURLHandler.On("Shorten", ts.URL, tt.url).Return(ts.URL+tt.res, tt.err)

			statusCode, respBody, header := testRequest(t, ts, http.MethodPost, "/api/shorten", []byte(jsonBody))

			mockedURLHandler.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, statusCode)
				return
			}

			assert.Equal(t, http.StatusCreated, statusCode)
			assert.Equal(t, tt.cType, header.Get("Content-Type"))
			assert.Equal(t, jsonResp, string(respBody))
		})
	}
}
