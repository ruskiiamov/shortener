package server

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testBaseURL       = "http://127.0.0.1:8080"
	testServerAddress = "127.0.0.1:8080"
	testSignKey       = "secret"
)

var mConverter *mockedConverter
var ts *httptest.Server

func init() {
	config := Config{
		BaseURL: testBaseURL,
		SignKey: testSignKey,
	}

	mConverter = new(mockedConverter)
	h := NewHandler(context.Background(), mConverter, chi.NewRouter(), config)

	ts = httptest.NewUnstartedServer(h)
	l, err := net.Listen("tcp", testServerAddress)
	if err != nil {
		panic(err)
	}
	ts.Listener.Close()
	ts.Listener = l

	ts.Start()
}

func TestGetUrl(t *testing.T) {
	tests := []struct {
		name    string
		encID   string
		res     *url.URL
		err     error
		wantErr bool
	}{
		{
			name:  "ok",
			encID: "1",
			res: &url.URL{
				EncodedID: "1",
				Original:  "http://shortener.com",
			},
			err:     nil,
			wantErr: false,
		},
		{
			name:    "not ok",
			encID:   "abc",
			res:     nil,
			err:     errors.New("wrong id"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mConverter.On("GetOriginal", mock.Anything, tt.encID).Return(tt.res, tt.err)

			statusCode, _, header := testRequest(t, ts, http.MethodGet, "/"+tt.encID, nil, nil)

			mConverter.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, statusCode)
				return
			}

			assert.Equal(t, http.StatusTemporaryRedirect, statusCode)
			assert.Equal(t, tt.res.Original, header.Get("Location"))
		})
	}
}

func TestAddURL(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		userID     string
		authCookie string
		res        *url.URL
		err        error
		want       string
		wantBody   bool
		status     int
	}{
		{
			name:       "ok",
			body:       "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res: &url.URL{
				EncodedID: "1",
				Original:  "http://shortener.com",
			},
			err:      nil,
			want:     ts.URL + "/1",
			wantBody: true,
			status:   201,
		},
		{
			name:       "not ok",
			body:       "shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        nil,
			err:        errors.New("wrong url"),
			want:       "wrong url\n",
			wantBody:   false,
			status:     500,
		},
		{
			name:       "duplicate",
			body:       "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        nil,
			err:        &url.ErrURLDuplicate{EncodedID: "4"},
			want:       ts.URL + "/4",
			wantBody:   true,
			status:     409,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mConverter.On("Shorten", mock.Anything, tt.userID, tt.body).Return(tt.res, tt.err).Once()

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, body, _ := testRequest(t, ts, http.MethodPost, "/", []byte(tt.body), cookie)

			mConverter.AssertExpectations(t)

			assert.Equal(t, tt.status, statusCode)
			assert.Equal(t, tt.want, string(body))
		})
	}
}

func TestAddURLFromJSON(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		userID     string
		authCookie string
		res        *url.URL
		cType      string
		err        error
		wantErr    bool
		status     int
		jsonResp   string
	}{
		{
			name:       "ok",
			url:        "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res: &url.URL{
				EncodedID: "1",
				Original:  "http://shortener.com",
			},
			cType:    "application/json",
			err:      nil,
			wantErr:  false,
			status:   201,
			jsonResp: `{"result":"http://127.0.0.1:8080/1"}`,
		},
		{
			name:       "not ok",
			url:        "shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        nil,
			cType:      "",
			err:        errors.New("wrong url"),
			wantErr:    true,
			status:     500,
			jsonResp:   "",
		},
		{
			name:       "duplicate",
			url:        "http://shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        nil,
			cType:      "application/json",
			err:        &url.ErrURLDuplicate{EncodedID: "4"},
			wantErr:    false,
			status:     409,
			jsonResp:   `{"result":"http://127.0.0.1:8080/4"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody := `{"url":"` + tt.url + `"}`

			mConverter.On("Shorten", mock.Anything, tt.userID, tt.url).Return(tt.res, tt.err).Once()

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodPost, "/api/shorten", []byte(jsonBody), cookie)

			mConverter.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, statusCode)
				return
			}

			assert.Equal(t, tt.status, statusCode)
			assert.Equal(t, tt.cType, header.Get("Content-Type"))
			assert.Equal(t, tt.jsonResp, string(respBody))
		})
	}
}

func TestAddURLBatch(t *testing.T) {
	tests := []struct {
		name       string
		authCookie string
		userID     string
		jsonBody   string
		originals  []string
		shortURLs  []url.URL
		respBody   string
	}{
		{
			name:       "ok",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			jsonBody:   `[{"correlation_id":"1","original_url":"http://shortener1.com"},{"correlation_id":"2","original_url":"http://shortener2.com"}]`,
			originals:  []string{"http://shortener1.com", "http://shortener2.com"},
			shortURLs: []url.URL{
				{EncodedID: "5", Original: "http://shortener1.com"},
				{EncodedID: "6", Original: "http://shortener2.com"},
			},
			respBody: `[{"correlation_id":"1","short_url":"http://127.0.0.1:8080/5"},{"correlation_id":"2","short_url":"http://127.0.0.1:8080/6"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mConverter.On("ShortenBatch", mock.Anything, tt.userID, tt.originals).Return(tt.shortURLs, nil)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodPost, "/api/shorten/batch", []byte(tt.jsonBody), cookie)

			assert.Equal(t, http.StatusCreated, statusCode)
			assert.Equal(t, "application/json", header.Get("Content-Type"))
			assert.JSONEq(t, tt.respBody, respBody)
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
					EncodedID: "1",
					Original:  "http://very-long-url/0",
				},
				{
					EncodedID: "2",
					Original:  "http://very-long-url/1",
				},
			},
			err:    nil,
			cType:  "application/json",
			status: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respData := []responseAll{}
			for _, item := range tt.res {
				respData = append(respData, responseAll{
					ShortURL:    ts.URL + "/" + item.EncodedID,
					OriginalURL: item.Original,
				})
			}
			jsonResp, _ := json.Marshal(respData)

			mConverter.On("GetAllByUser", mock.Anything, tt.userID).Return(tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodGet, "/api/user/urls", nil, cookie)

			mConverter.AssertExpectations(t)

			assert.Equal(t, tt.status, statusCode)
			assert.Equal(t, tt.cType, header.Get("Content-Type"))
			assert.Equal(t, string(jsonResp), respBody)
		})
	}
}

func TestDeleteURLBatch(t *testing.T) {
	authCookie := "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ="
	cookie := &http.Cookie{Name: authCookieName, Value: authCookie}

	jsonBody := `["1","2","5"]`

	statusCode, _, _ := testRequest(t, ts, http.MethodDelete, "/api/user/urls", []byte(jsonBody), cookie)

	assert.Equal(t, 202, statusCode)
}

func TestPingDB(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "ok",
			err:    nil,
			status: 200,
		},
		{
			name:   "not ok",
			err:    errors.New("some error"),
			status: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCookie := "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ="
			cookie := &http.Cookie{Name: authCookieName, Value: authCookie}

			mConverter.On("PingKeeper", mock.Anything).Return(tt.err).Once()

			statusCode, _, _ := testRequest(t, ts, http.MethodGet, "/ping", nil, cookie)

			assert.Equal(t, tt.status, statusCode)
		})
	}
}
