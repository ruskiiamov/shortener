package server

import "github.com/stretchr/testify/mock"

type MockedURLHandler struct {
	mock.Mock
}

func (m *MockedURLHandler) Shorten(host, url string) (string, error) {
	args := m.Called(host, url)
	return args.String(0), args.Error(1)
}

func (m *MockedURLHandler) GetOriginal(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}
