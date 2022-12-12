package url

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		res     string
		err     error
	}{
		{
			name:    "ok",
			args:    args{url: "http://shortener.com"},
			want:    "http://localhost:8080/0",
			wantErr: false,
			res:     "0",
			err:     nil,
		},
		{
			name:    "not ok",
			args:    args{url: "shortener.com"},
			want:    "http://localhost:8080/0",
			wantErr: true,
			res:     "",
			err:     errors.New("wrong url"),
		},
	}

	mockedDataKeeper := new(MockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("Add", OriginalURL{
				URL: tt.args.url,
			}).Return(tt.res, tt.err)

			c := NewConverter(mockedDataKeeper, "http://localhost:8080")

			got, err := c.Shorten(tt.args.url)

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
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		res     *OriginalURL
		err     error
	}{
		{
			name:    "ok",
			args:    args{id: "0"},
			want:    "http://shortener.com",
			wantErr: false,
			res:     &OriginalURL{URL: "http://shortener.com"},
			err:     nil,
		},
		{
			name:    "not ok",
			args:    args{id: "abc"},
			want:    "http://shortener.com",
			wantErr: true,
			res:     nil,
			err:     errors.New("wrong id"),
		},
	}

	mockedDataKeeper := new(MockedDataKeeper)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedDataKeeper.On("Get", tt.args.id).Return(tt.res, tt.err)

			c := &converter{
				dataKeeper: mockedDataKeeper,
			}

			got, err := c.GetOriginal(tt.args.id)

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
