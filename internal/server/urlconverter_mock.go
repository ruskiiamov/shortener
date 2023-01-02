package server

import (
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/mock"
)

type mockedConverter struct {
	mock.Mock
}

func (m *mockedConverter) Shorten(url, userID string) (string, error) {
	args := m.Called(url, userID)
	return args.String(0), args.Error(1)
}

func (m *mockedConverter) GetOriginal(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *mockedConverter) GetAll(userID string) ([]url.URL, error) {
	args := m.Called(userID)
	return args.Get(0).([]url.URL), args.Error(1)
}
