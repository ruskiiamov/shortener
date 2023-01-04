package url

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		userID  string
		want    string
		wantErr bool
		res     string
		err     error
	}{
		{
			name:    "ok",
			url:     "http://shortener.com",
			userID:  "7b6def87-f3dc-4036-bda2-3a6ca1298ef5",
			want:    "http://localhost:8080/0",
			wantErr: false,
			res:     "0",
			err:     nil,
		},
		{
			name:    "not ok",
			url:     "shortener.com",
			userID:  "cdc04719-2852-456f-bacb-2ce370678013",
			want:    "http://localhost:8080/0",
			wantErr: true,
			res:     "",
			err:     errors.New("wrong url"),
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				mockedDataKeeper.On("Add", OriginalURL{
					URL:    tt.url,
					UserID: tt.userID,
				}).Return(tt.res, tt.err)
			}

			c := NewConverter(mockedDataKeeper, "http://localhost:8080")

			got, err := c.Shorten(tt.url, tt.userID)

			mockedDataKeeper.AssertExpectations(t)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetOriginal(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
		res     *OriginalURL
		err     error
	}{
		{
			name:    "ok",
			id:      "0",
			want:    "http://shortener.com",
			wantErr: false,
			res:     &OriginalURL{URL: "http://shortener.com"},
			err:     nil,
		},
		{
			name:    "not ok",
			id:      "abc",
			want:    "http://shortener.com",
			wantErr: true,
			res:     nil,
			err:     errors.New("wrong id"),
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("Get", tt.id).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper, "http://localhost:8080")

			got, err := c.GetOriginal(tt.id)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAll(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		want    []URL
		wantErr bool
		res     []OriginalURL
		err     error
	}{
		{
			name:   "ok",
			userID: "21f923fc-cbbf-4fb1-a05c-21933d307be2",
			want: []URL{
				{
					ShortURL:    "http://localhost:8080/1",
					OriginalURL: "http://shortener.com",
				},
				{
					ShortURL:    "http://localhost:8080/3",
					OriginalURL: "http://shortener.ru",
				},
			},
			wantErr: false,
			res: []OriginalURL{
				{
					ID:     "1",
					URL:    "http://shortener.com",
					UserID: "21f923fc-cbbf-4fb1-a05c-21933d307be2",
				},
				{
					ID:     "3",
					URL:    "http://shortener.ru",
					UserID: "21f923fc-cbbf-4fb1-a05c-21933d307be2",
				},
			},
			err: nil,
		},
	}

	mockedDataKeeper := new(mockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("GetAllByUser", tt.userID).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper, "http://localhost:8080")

			got, err := c.GetAll(tt.userID)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
