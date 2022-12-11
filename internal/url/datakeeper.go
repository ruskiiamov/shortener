package url

import (
	"github.com/stretchr/testify/mock"
)

type MockedDataKeeper struct {
	mock.Mock
}

func (m *MockedDataKeeper) Add(url OriginalURL) (id string, err error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockedDataKeeper) Get(id string) (*OriginalURL, error) {
	args := m.Called(id)
	return args.Get(0).(*OriginalURL), args.Error(1)
}
