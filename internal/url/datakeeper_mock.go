package url

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedDataKeeper struct {
	mock.Mock
}

func (m *mockedDataKeeper) Add(ctx context.Context, userID, original string) (int, error) {
	args := m.Called(ctx, userID, original)
	return args.Int(0), args.Error(1)
}

func (m *mockedDataKeeper) AddBatch(ctx context.Context, userID string, originals []string) (map[string]int, error) {
	args := m.Called(ctx, userID, originals)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockedDataKeeper) Get(ctx context.Context, id int) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *mockedDataKeeper) GetAllByUser(ctx context.Context, userID string) (map[string]int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockedDataKeeper) DeleteBatch(ctx context.Context, batch map[string][]int) error {
	args := m.Called(ctx, batch)
	return args.Error(0)
}

func (m *mockedDataKeeper) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockedDataKeeper) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
