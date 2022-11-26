package repo

import (
	"testing"

	"github.com/ruskiiamov/shortener/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	type fields struct {
		s []string
	}
	type args struct {
		url entity.ShortenedURL
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "ok",
			fields: fields{s: []string{}},
			args: args{
				url: entity.ShortenedURL{
					OriginalURL: "http://shortener.com",
				},
			},
			wantErr: false,
		},
		{
			name:   "wrong url",
			fields: fields{s: []string{}},
			args: args{
				url: entity.ShortenedURL{
					OriginalURL: "shortener.com",
				},
			},
			wantErr: true,
		},
		{
			name:   "repeat url",
			fields: fields{s: []string{"http://shortener.com"}},
			args: args{
				url: entity.ShortenedURL{
					OriginalURL: "http://shortener.com",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerSlice{
				s: tt.fields.s,
			}

			id, err := s.Add(tt.args.url)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.NotEmpty(t, id)
		})
	}
}

func TestGet(t *testing.T) {
	type fields struct {
		s []string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.ShortenedURL
		wantErr bool
	}{
		{
			name:   "ok",
			fields: fields{s: []string{"http://shortener.com"}},
			args:   args{id: "0"},
			want: &entity.ShortenedURL{
				ID:          "0",
				OriginalURL: "http://shortener.com",
			},
			wantErr: false,
		},
		{
			name:   "not int id",
			fields: fields{s: []string{"http://shortener.com"}},
			args:   args{id: "abc"},
			want: &entity.ShortenedURL{
				ID:          "abc",
				OriginalURL: "http://shortener.com",
			},
			wantErr: true,
		},
		{
			name:   "negative id",
			fields: fields{s: []string{"http://shortener.com"}},
			args:   args{id: "-2"},
			want: &entity.ShortenedURL{
				ID:          "-2",
				OriginalURL: "http://shortener.com",
			},
			wantErr: true,
		},
		{
			name:   "too big id",
			fields: fields{s: []string{"http://shortener.com", "http://shortener.com/info"}},
			args:   args{id: "2"},
			want: &entity.ShortenedURL{
				ID:          "2",
				OriginalURL: "http://shortener.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortenerSlice{
				s: tt.fields.s,
			}

			got, err := s.Get(tt.args.id)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
