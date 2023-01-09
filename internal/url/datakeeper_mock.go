package url

import (
	"github.com/stretchr/testify/mock"
)

type mockedDataKeeper struct {
	mock.Mock
}

func (m *mockedDataKeeper) Add(userID, original string) (int, error) {
	args := m.Called(userID, original)
	return args.Int(0), args.Error(1)
}

func (m *mockedDataKeeper) AddBatch(userID string, originals []string) (map[string]int, error) {
	args := m.Called(userID, originals)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockedDataKeeper) Get(id int) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *mockedDataKeeper) GetAllByUser(userID string) (map[string]int, error) {
	args := m.Called(userID)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockedDataKeeper) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockedDataKeeper) Close() error {
	args := m.Called()
	return args.Error(0)
}
