package server

import (
	"context"

	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/mock"
)

type mockedConverter struct {
	mock.Mock
}

// Shorten is mocked method.
func (m *mockedConverter) Shorten(ctx context.Context, userID, original string) (*url.URL, error) {
	args := m.Called(ctx, userID, original)
	return args.Get(0).(*url.URL), args.Error(1)
}

// ShortenBatch is mocked method.
func (m *mockedConverter) ShortenBatch(ctx context.Context, userID string, originals []string) ([]url.URL, error) {
	args := m.Called(ctx, userID, originals)
	return args.Get(0).([]url.URL), args.Error(1)
}

// GetOriginal is mocked method.
func (m *mockedConverter) GetOriginal(ctx context.Context, encodedID string) (*url.URL, error) {
	args := m.Called(ctx, encodedID)
	return args.Get(0).(*url.URL), args.Error(1)
}

// GetAllByUser is mocked method.
func (m *mockedConverter) GetAllByUser(ctx context.Context, userID string) ([]url.URL, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]url.URL), args.Error(1)
}

// RemoveBatch is mocked method.
func (m *mockedConverter) RemoveBatch(ctx context.Context, batch map[string][]string) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

// PingKeeper is mocked method.
func (m *mockedConverter) PingKeeper(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockedConverter) GetStats(ctx context.Context) (urls, users int, err error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}
