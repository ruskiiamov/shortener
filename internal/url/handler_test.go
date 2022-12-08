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

	mockedStorage := new(MockedStorage)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedStorage.On("Add", OriginalURL{
				URL: tt.args.url,
			}).Return(tt.res, tt.err)

			h := &handler{
				storage: mockedStorage,
			}

			got, err := h.Shorten("http://localhost:8080", tt.args.url)

			mockedStorage.AssertExpectations(t)

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

	mockedStorage := new(MockedStorage)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedStorage.On("Get", tt.args.id).Return(tt.res, tt.err)

			h := &handler{
				storage: mockedStorage,
			}

			got, err := h.GetOriginal(tt.args.id)

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
