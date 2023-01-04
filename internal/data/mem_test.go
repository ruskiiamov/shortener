package data

import (
	"testing"

	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/assert"
)

func TestMemAdd(t *testing.T) {
	tests := []struct {
		name    string
		keeper  memKeeper
		url     url.OriginalURL
		wantErr bool
	}{
		{
			name:   "ok",
			keeper: memKeeper{},
			url: url.OriginalURL{
				URL:    "http://shortener.com",
				UserID: "1770aae6-caaf-4578-b27e-ffa967927a1b",
			},
			wantErr: false,
		},
		{
			name: "repeat url",
			keeper: memKeeper{url.OriginalURL{
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			}},
			url: url.OriginalURL{
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := tt.keeper.Add(tt.url)

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
	tests := []struct {
		name    string
		keeper  memKeeper
		id      string
		want    *url.OriginalURL
		wantErr bool
	}{
		{
			name: "ok",
			keeper: memKeeper{url.OriginalURL{
				ID:     "0",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			}},
			id: "0",
			want: &url.OriginalURL{
				ID:     "0",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			},
			wantErr: false,
		},
		{
			name: "not int id",
			keeper: memKeeper{url.OriginalURL{
				ID:     "0",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			}},
			id: "abc",
			want: &url.OriginalURL{
				ID:     "abc",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			},
			wantErr: true,
		},
		{
			name: "negative id",
			keeper: memKeeper{url.OriginalURL{
				ID:     "0",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			}},
			id: "-2",
			want: &url.OriginalURL{
				ID:     "-2",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			},
			wantErr: true,
		},
		{
			name: "too big id",
			keeper: memKeeper{
				url.OriginalURL{
					ID:     "0",
					URL:    "http://shortener.com",
					UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
				},
				url.OriginalURL{
					ID:     "1",
					URL:    "http://shortener.com/info",
					UserID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
				},
			},
			id: "2",
			want: &url.OriginalURL{
				ID:     "2",
				URL:    "http://shortener.com",
				UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keeper.Get(tt.id)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAllByUser(t *testing.T) {
	tests := []struct {
		name    string
		keeper  memKeeper
		userID  string
		want    []url.OriginalURL
		wantErr bool
	}{
		{
			name: "ok",
			keeper: memKeeper{
				url.OriginalURL{
					ID:     "0",
					URL:    "http://shortener.com",
					UserID: "c7cbe16d-034e-40b9-a2a5-e936851c4282",
				},
				url.OriginalURL{
					ID:     "1",
					URL:    "http://shortener.com/info",
					UserID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
				},
				url.OriginalURL{
					ID:     "2",
					URL:    "http://shortener.com/stat",
					UserID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
				},
			},
			userID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
			want: []url.OriginalURL{
				{
					ID:     "1",
					URL:    "http://shortener.com/info",
					UserID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
				},
				{
					ID:     "2",
					URL:    "http://shortener.com/stat",
					UserID: "b01ad148-d4da-4b08-9c75-9eb66899119f",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.keeper.GetAllByUser(tt.userID)

			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
