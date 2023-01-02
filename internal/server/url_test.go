package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/assert"
)

const testSignKey = "secret"

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

	mockedConverter := new(mockedConverter)
	h := NewHandler(mockedConverter, chi.NewRouter(), nil, testSignKey)
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedConverter.On("GetOriginal", tt.id).Return(tt.res, tt.err)

			statusCode, _, header := testRequest(t, ts, http.MethodGet, tt.path, nil, nil)

			mockedConverter.AssertExpectations(t)

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
		name       string
		body       string
		userID     string
		authCookie string
		res        string
		err        error
		wantErr    bool
	}{
		{
			name:       "ok",
			body:       "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        "/0",
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "not ok",
			body:       "shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        "",
			err:        errors.New("wrong url"),
			wantErr:    true,
		},
	}

	mockedURLHandler := new(mockedConverter)
	h := NewHandler(mockedURLHandler, chi.NewRouter(), nil, testSignKey)
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedURLHandler.On("Shorten", tt.body, tt.userID).Return(ts.URL+tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, body, _ := testRequest(t, ts, http.MethodPost, "/", []byte(tt.body), cookie)

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
		name       string
		url        string
		userID     string
		authCookie string
		res        string
		cType      string
		err        error
		wantErr    bool
	}{
		{
			name:       "ok",
			url:        "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        "/0",
			cType:      "application/json",
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "not ok",
			url:        "shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        "",
			cType:      "",
			err:        errors.New("wrong url"),
			wantErr:    true,
		},
	}

	mockedURLHandler := new(mockedConverter)
	h := NewHandler(mockedURLHandler, chi.NewRouter(), nil, testSignKey)
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody := `{"url":"` + tt.url + `"}`
			jsonResp := `{"result":"` + ts.URL + tt.res + `"}`
			mockedURLHandler.On("Shorten", tt.url, tt.userID).Return(ts.URL+tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodPost, "/api/shorten", []byte(jsonBody), cookie)

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

func TestGetAllURL(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		authCookie string
		res        []url.URL
		err        error
		cType      string
		status     int
	}{
		{
			name:       "ok",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res: []url.URL{
				{
					ShortURL:    "/0",
					OriginalURL: "http://very-long-url/0",
				},
				{
					ShortURL:    "/1",
					OriginalURL: "http://very-long-url/1",
				},
			},
			err:    nil,
			cType:  "application/json",
			status: http.StatusOK,
		},
		{
			name:       "empty",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        []url.URL{},
			err:        nil,
			cType:      "",
			status:     http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedURLHandler := new(mockedConverter)
			h := NewHandler(mockedURLHandler, chi.NewRouter(), nil, testSignKey)
			ts := httptest.NewServer(h)
			defer ts.Close()

			for i := range tt.res {
				tt.res[i].ShortURL = ts.URL + tt.res[i].ShortURL
			}

			var jsonResp []byte

			if len(tt.res) != 0 {
				jsonResp, _ = json.Marshal(tt.res)
			} else {
				jsonResp = []byte{}
			}

			mockedURLHandler.On("GetAll", tt.userID).Return(tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil, cookie)

			mockedURLHandler.AssertExpectations(t)

			assert.Equal(t, tt.status, statusCode)
			assert.Equal(t, tt.cType, header.Get("Content-Type"))
			assert.Equal(t, string(jsonResp), respBody)
		})
	}
}
