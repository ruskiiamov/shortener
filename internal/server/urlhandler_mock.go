package server

import "github.com/stretchr/testify/mock"

type MockedConverter struct {
	mock.Mock
}

func (m *MockedConverter) Shorten(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockedConverter) GetOriginal(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}
