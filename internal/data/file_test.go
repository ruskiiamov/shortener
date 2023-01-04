package data

import (
	"os"
	"testing"

	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/stretchr/testify/assert"
)

func TestFileAdd(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		fileData   string
		fileExists bool
		url        url.OriginalURL
		want       string
	}{
		{
			name:       "ok",
			filePath:   "test_data_file",
			fileExists: false,
			fileData:   "",
			url: url.OriginalURL{
				URL:    "https://very-long-url.com",
				UserID: "e01a511c-bfaa-4f4e-80f9-e3f07f8664ee",
			},
			want: "0",
		},
		{
			name:       "repeat url",
			filePath:   "test_data_file",
			fileExists: true,
			fileData: `{"id":"0","url":"https://very-long-url-0.com","user_id":"e01a511c-bfaa-4f4e-80f9-e3f07f8664ee"}` +
				"\n" + `{"id":"1","url":"https://very-long-url-1.com","user_id":"1930751f-6252-4892-95de-ef528eabeb39"}`,
			url: url.OriginalURL{
				URL:    "https://very-long-url-1.com",
				UserID: "1930751f-6252-4892-95de-ef528eabeb39",
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileExists {
				file, _ := os.Create(tt.filePath)
				file.Write([]byte(tt.fileData))
				file.Close()
			}

			keeper, _ := NewKeeper("", tt.filePath)
			got, err := keeper.Add(tt.url)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			if tt.fileExists {
				os.Remove(tt.filePath)
			}
		})
	}
}

func TestFileGet(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		fileData string
		id       string
		want     *url.OriginalURL
		wantErr  bool
	}{
		{
			name:     "ok",
			filePath: "test_data_file",
			fileData: `{"id":"0","url":"https://very-long-url.com","user_id":"f940c007-496f-4507-b41f-3cd43d7e7286"}`,
			id:       "0",
			want: &url.OriginalURL{
				ID:     "0",
				URL:    "https://very-long-url.com",
				UserID: "f940c007-496f-4507-b41f-3cd43d7e7286",
			},
			wantErr: false,
		},
		{
			name:     "not found",
			filePath: "test_data_file",
			fileData: `{"id":"0","url":"https://very-long-url-0.com","user_id":"f940c007-496f-4507-b41f-3cd43d7e7286"}` +
				"\n" + `{"id":"1","url":"https://very-long-url-1.com","user_id":"e01a511c-bfaa-4f4e-80f9-e3f07f8664ee"}`,
			id:      "2",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.filePath)
			file.Write([]byte(tt.fileData))
			file.Close()

			keeper, _ := NewKeeper("", tt.filePath)
			got, err := keeper.Get(tt.id)

			os.Remove(tt.filePath)

			if tt.wantErr {
				assert.Empty(t, got)
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileGetAllByUser(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		fileData string
		userID   string
		want     []url.OriginalURL
		wantErr  bool
	}{
		{
			name:     "ok",
			filePath: "test_data_file",
			fileData: `{"id":"0","url":"https://very-long-url-0.com","user_id":"f940c007-496f-4507-b41f-3cd43d7e7286"}` +
				"\n" + `{"id":"1","url":"https://very-long-url-1.com","user_id":"e01a511c-bfaa-4f4e-80f9-e3f07f8664ee"}` +
				"\n" + `{"id":"2","url":"https://very-long-url-2.com","user_id":"f940c007-496f-4507-b41f-3cd43d7e7286"}`,
			userID: "f940c007-496f-4507-b41f-3cd43d7e7286",
			want: []url.OriginalURL{
				{
					ID:     "0",
					URL:    "https://very-long-url-0.com",
					UserID: "f940c007-496f-4507-b41f-3cd43d7e7286",
				},
				{
					ID:     "2",
					URL:    "https://very-long-url-2.com",
					UserID: "f940c007-496f-4507-b41f-3cd43d7e7286",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.filePath)
			file.Write([]byte(tt.fileData))
			file.Close()

			keeper, _ := NewKeeper("", tt.filePath)
			got, err := keeper.GetAllByUser(tt.userID)

			os.Remove(tt.filePath)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
