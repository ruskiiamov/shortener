package url

import (
	"github.com/stretchr/testify/mock"
)

type mockedDataKeeper struct {
	mock.Mock
}

func (m *mockedDataKeeper) Add(url OriginalURL) (id string, err error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *mockedDataKeeper) Get(id string) (*OriginalURL, error) {
	args := m.Called(id)
	return args.Get(0).(*OriginalURL), args.Error(1)
}

func (m *mockedDataKeeper) GetAllByUser(userID string) ([]OriginalURL, error) {
	args := m.Called(userID)
	return args.Get(0).([]OriginalURL), args.Error(1)
}
