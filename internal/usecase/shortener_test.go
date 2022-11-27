package usecase

import (
	"errors"
	"testing"

	"github.com/ruskiiamov/shortener/internal/entity"
	"github.com/ruskiiamov/shortener/internal/usecase/repo"
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

	mockedShortenerRepo := new(repo.MockedShortenerRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedShortenerRepo.On("Add", entity.ShortenedURL{
				OriginalURL: tt.args.url,
			}).Return(tt.res, tt.err)

			uc := &shortenerUseCase{
				repo: mockedShortenerRepo,
			}

			got, err := uc.Shorten("http://localhost:8080", tt.args.url)

			mockedShortenerRepo.AssertExpectations(t)

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
		res     *entity.ShortenedURL
		err     error
	}{
		{
			name:    "ok",
			args:    args{id: "0"},
			want:    "http://shortener.com",
			wantErr: false,
			res:     &entity.ShortenedURL{OriginalURL: "http://shortener.com"},
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

	mockedShortenerRepo := new(repo.MockedShortenerRepo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedShortenerRepo.On("Get", tt.args.id).Return(tt.res, tt.err)

			uc := &shortenerUseCase{
				repo: mockedShortenerRepo,
			}

			got, err := uc.GetOriginal(tt.args.id)

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
