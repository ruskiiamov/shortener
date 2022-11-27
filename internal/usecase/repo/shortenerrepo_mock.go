package repo

import (
	"github.com/ruskiiamov/shortener/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockedShortenerRepo struct {
	mock.Mock
}

func (m *MockedShortenerRepo) Add(shortenedURL entity.ShortenedURL) (id string, err error) {
	args := m.Called(shortenedURL)
	return args.String(0), args.Error(1)
}

func (m *MockedShortenerRepo) Get(id string) (*entity.ShortenedURL, error) {
	args := m.Called(id)
	return args.Get(0).(*entity.ShortenedURL), args.Error(1)
}
