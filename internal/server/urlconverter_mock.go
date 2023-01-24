package server

import (
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/mock"
)

type mockedConverter struct {
	mock.Mock
}

func (m *mockedConverter) Shorten(userID, original string) (*url.URL, error) {
	args := m.Called(userID, original)
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *mockedConverter) ShortenBatch(userID string, originals []string) ([]url.URL, error) {
	args := m.Called(userID, originals)
	return args.Get(0).([]url.URL), args.Error(1)
}

func (m *mockedConverter) GetOriginal(encodedID string) (*url.URL, error) {
	args := m.Called(encodedID)
	return args.Get(0).(*url.URL), args.Error(1)
}

func (m *mockedConverter) GetAllByUser(userID string) ([]url.URL, error) {
	args := m.Called(userID)
	return args.Get(0).([]url.URL), args.Error(1)
}

func (m *mockedConverter) RemoveBatch(batch map[string][]string) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *mockedConverter) PingKeeper() error {
	args := m.Called()
	return args.Error(0)
}
