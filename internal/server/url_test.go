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
		wantErr    bool
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
			err:     nil,
			want:    ts.URL + "/1",
			wantErr: false,
		},
		{
			name:       "not ok",
			body:       "shortener.com",
			userID:     "cfb31f30-efa9-4244-b1d6-e04c8438771d",
			authCookie: "XlBVspVMtREN3fydYOxHRdxJKff1Emw3UwLB5RgQrj9jZmIzMWYzMC1lZmE5LTQyNDQtYjFkNi1lMDRjODQzODc3MWQ=",
			res:        nil,
			err:        errors.New("wrong url"),
			want:       "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mConverter.On("Shorten", mock.Anything, tt.userID, tt.body).Return(tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, body, _ := testRequest(t, ts, http.MethodPost, "/", []byte(tt.body), cookie)

			mConverter.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusInternalServerError, statusCode)
				return
			}

			assert.Equal(t, http.StatusCreated, statusCode)
			assert.Equal(t, ts.URL+"/"+tt.res.EncodedID, string(body))
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
			cType:   "application/json",
			err:     nil,
			wantErr: false,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody := `{"url":"` + tt.url + `"}`
			var jsonResp string
			if tt.res != nil {
				jsonResp = `{"result":"` + ts.URL + "/" + tt.res.EncodedID + `"}`
			} else {
				jsonResp = ""
			}

			mConverter.On("Shorten", mock.Anything, tt.userID, tt.url).Return(tt.res, tt.err)

			cookie := &http.Cookie{Name: authCookieName, Value: tt.authCookie}

			statusCode, respBody, header := testRequest(t, ts, http.MethodPost, "/api/shorten", []byte(jsonBody), cookie)

			mConverter.AssertExpectations(t)

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
