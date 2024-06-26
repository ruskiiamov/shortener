package server

import "github.com/stretchr/testify/mock"

type mockedUserAuth struct {
	mock.Mock
}

// CreateUser is mocked method.
func (m *mockedUserAuth) CreateUser() (userID, token string, err error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

// GetUserID is mocked method.
func (m *mockedUserAuth) GetUserID(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}
