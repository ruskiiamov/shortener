package data

import (
	"testing"

	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/assert"
)

func TestMemAdd(t *testing.T) {
	type args struct {
		url url.OriginalURL
	}
	tests := []struct {
		name    string
		keeper  dataMemKeeper
		args    args
		wantErr bool
	}{
		{
			name:   "ok",
			keeper: dataMemKeeper([]string{}),
			args: args{
				url: url.OriginalURL{
					URL: "http://shortener.com",
				},
			},
			wantErr: false,
		},
		{
			name:   "repeat url",
			keeper: dataMemKeeper([]string{"http://shortener.com"}),
			args: args{
				url: url.OriginalURL{
					URL: "http://shortener.com",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tt.keeper.Add(tt.args.url)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.NotEmpty(t, id)
		})
	}
}

func TestMemGet(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		keeper  dataMemKeeper
		args    args
		want    *url.OriginalURL
		wantErr bool
	}{
		{
			name:   "ok",
			keeper: dataMemKeeper([]string{"http://shortener.com"}),
			args:   args{id: "0"},
			want: &url.OriginalURL{
				ID:  "0",
				URL: "http://shortener.com",
			},
			wantErr: false,
		},
		{
			name:   "not int id",
			keeper: dataMemKeeper([]string{"http://shortener.com"}),
			args:   args{id: "abc"},
			want: &url.OriginalURL{
				ID:  "abc",
				URL: "http://shortener.com",
			},
			wantErr: true,
		},
		{
			name:   "negative id",
			keeper: dataMemKeeper([]string{"http://shortener.com"}),
			args:   args{id: "-2"},
			want: &url.OriginalURL{
				ID:  "-2",
				URL: "http://shortener.com",
			},
			wantErr: true,
		},
		{
			name:   "too big id",
			keeper: dataMemKeeper([]string{"http://shortener.com", "http://shortener.com/info"}),
			args:   args{id: "2"},
			want: &url.OriginalURL{
				ID:  "2",
				URL: "http://shortener.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keeper.Get(tt.args.id)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
