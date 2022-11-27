package usecase

import "github.com/stretchr/testify/mock"

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
