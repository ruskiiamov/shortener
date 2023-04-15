package url

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		userID   string
		want     string
		wantErr  bool
		res      int
		err      error
		checkErr bool
		keeper   bool
	}{
		{
			name:     "ok",
			url:      "http://shortener.com",
			userID:   "7b6def87-f3dc-4036-bda2-3a6ca1298ef5",
			want:     "1",
			wantErr:  false,
			res:      1,
			err:      nil,
			checkErr: false,
			keeper:   true,
		},
		{
			name:    "duplicate",
			url:     "http://shortener.com",
			userID:  "7b6def87-f3dc-4036-bda2-3a6ca1298ef5",
			want:    "1",
			wantErr: true,
			res:     0,
			err: &ErrURLDuplicate{
				ID:  1,
				URL: "http://shortener.com",
			},
			checkErr: true,
			keeper:   true,
		},
		{
			name:     "not correct url",
			url:      "shortener.com",
			want:     "1",
			wantErr:  true,
			res:      0,
			err:      errors.New("test"),
			checkErr: false,
			keeper:   false,
		},
		{
			name:     "keeper error",
			url:      "http://shortener.com",
			want:     "1",
			wantErr:  true,
			res:      0,
			err:      errors.New("test"),
			checkErr: false,
			keeper:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper := new(mockedDataKeeper)
			mockedDataKeeper.On("Add", context.Background(), tt.userID, tt.url).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper)
			got, err := c.Shorten(context.Background(), tt.userID, tt.url)

			if tt.keeper {
				mockedDataKeeper.AssertExpectations(t)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr {
					assert.ErrorIs(t, err, tt.err)
				}
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got.EncodedID)
			assert.Equal(t, tt.url, got.Original)
		})
	}
}

func TestShortenBatch(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		originals []string
		res       map[string]int
		err       error
		want      []URL
		wantErr   bool
	}{
		{
			name:      "ok",
			userID:    "7b6def87-f3dc-4036-bda2-3a6ca1298ef5",
			originals: []string{"https://shortener.com", "https://shortener2.ru"},
			res:       map[string]int{"https://shortener.com": 1, "https://shortener2.ru": 2},
			err:       nil,
			want: []URL{
				{
					EncodedID: "1",
					Original:  "https://shortener.com",
				},
				{
					EncodedID: "2",
					Original:  "https://shortener2.ru",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper := new(mockedDataKeeper)
			mockedDataKeeper.On("AddBatch", context.Background(), tt.userID, tt.originals).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper)
			got, err := c.ShortenBatch(context.Background(), tt.userID, tt.originals)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestGetOriginal(t *testing.T) {
	tests := []struct {
		name    string
		encID   string
		id      int
		want    string
		wantErr bool
		res     string
		err     error
	}{
		{
			name:    "ok",
			encID:   "1",
			id:      1,
			want:    "http://shortener.com",
			wantErr: false,
			res:     "http://shortener.com",
			err:     nil,
		},
		{
			name:    "not ok",
			encID:   "0",
			id:      0,
			want:    "http://shortener.com",
			wantErr: true,
			res:     "",
			err:     errors.New("wrong id"),
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("Get", context.Background(), tt.id).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper)

			got, err := c.GetOriginal(context.Background(), tt.encID)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.encID, got.EncodedID)
			assert.Equal(t, tt.want, got.Original)
		})
	}
}

func TestGetAllByUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		want    []URL
		wantErr bool
		res     map[string]int
		err     error
	}{
		{
			name:   "ok",
			userID: "21f923fc-cbbf-4fb1-a05c-21933d307be2",
			want: []URL{
				{
					EncodedID: "1",
					Original:  "http://shortener.com",
				},
				{
					EncodedID: "3",
					Original:  "http://shortener.ru",
				},
			},
			wantErr: false,
			res:     map[string]int{"http://shortener.com": 1, "http://shortener.ru": 3},
			err:     nil,
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("GetAllByUser", context.Background(), tt.userID).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper)

			got, err := c.GetAllByUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
				return
			}

			assert.Nil(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestRemoveBatch(t *testing.T) {
	tests := []struct {
		name         string
		batch        map[string][]string
		decodedBatch map[string][]int
		wantErr      bool
		dataErr      error
	}{
		{
			name: "ok",
			batch: map[string][]string{
				"21f923fc-cbbf-4fb1-a05c-21933d307be2": {"1", "3"},
			},
			decodedBatch: map[string][]int{
				"21f923fc-cbbf-4fb1-a05c-21933d307be2": {1, 3},
			},
			wantErr: false,
			dataErr: nil,
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("DeleteBatch", context.Background(), tt.decodedBatch).Return(tt.dataErr).Once()
			c := NewConverter(mockedDataKeeper)

			err := c.RemoveBatch(context.Background(), tt.batch)

			mockedDataKeeper.AssertExpectations(t)

			if tt.wantErr {
				assert.Error(t, err)
			}

			assert.NoError(t, err)
		})
	}
}

func TestGetStats(t *testing.T) {
	tests := []struct {
		name  string
		urls  int
		users int
		err   error
	}{
		{
			name:  "ok",
			urls:  123,
			users: 98,
			err:   nil,
		},
		{
			name:  "error",
			urls:  0,
			users: 0,
			err:   errors.New("test error"),
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run("ok", func(t *testing.T) {
			mockedDataKeeper.On("GetStats", context.Background()).Return(tt.urls, tt.users, tt.err).Once()
			c := NewConverter(mockedDataKeeper)

			urls, users, err := c.GetStats(context.Background())

			mockedDataKeeper.AssertExpectations(t)

			if tt.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.urls, urls)
			assert.Equal(t, tt.users, users)
		})
	}
}
