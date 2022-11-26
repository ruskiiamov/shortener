package controller

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedShortener struct {
	mock.Mock
}

func (m *MockedShortener) Shorten(host, url string) (string, error) {
	args := m.Called(host, url)
	return args.String(0), args.Error(1)
}

func (m *MockedShortener) GetOriginal(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		id          string
		res         string
		err         error
		wantErr     bool
	}{
		{
			name:        "ok",
			host:        "/0",
			id:          "0",
			res:         "http://shortener",
			err:         nil,
			wantErr:     false,
		},
		{
			name:        "not ok",
			host:        "/abc",
			id:          "abc",
			res:         "",
			err:         errors.New("wrong id"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.host, nil)
			w := httptest.NewRecorder()

			mockedShortener := new(MockedShortener)
			mockedShortener.On("GetOriginal", tt.id).Return(tt.res, tt.err)

			h := http.HandlerFunc(newShortenerHandler(mockedShortener))
			h.ServeHTTP(w, request)

			res := w.Result()

			mockedShortener.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				return
			}

			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			assert.Equal(t, tt.res, res.Header.Get("Location"))
		})
	}
}

func TestPost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		body    string
		res     string
		err     error
		wantErr bool
	}{
		{
			name:    "ok",
			host:    "http://localhost:8080",
			body:    "http://shortener.com",
			res:     "http://localhost:8080/0",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "not ok",
			host:    "http://localhost:8080",
			body:    "shortener.com",
			res:     "",
			err:     errors.New("wrong url"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.host, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			mockedShortener := new(MockedShortener)
			mockedShortener.On("Shorten", tt.host, tt.body).Return(tt.res, tt.err)

			h := http.HandlerFunc(newShortenerHandler(mockedShortener))
			h.ServeHTTP(w, request)

			res := w.Result()

			mockedShortener.AssertExpectations(t)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, res.StatusCode)
				return
			}

			assert.Equal(t, http.StatusCreated, res.StatusCode)
			body, err := io.ReadAll(res.Body)
			if err != nil {
				assert.Fail(t, "body error: "+err.Error())
				return
			}
			defer res.Body.Close()
			assert.Equal(t, tt.res, string(body))
		})
	}
}
