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
		keeper     url.DataKeeper
		url        url.OriginalURL
		want       string
	}{
		{
			name:       "ok",
			filePath:   "test_data_file",
			fileExists: false,
			fileData:   "",
			keeper:     NewKeeper("test_data_file"),
			url: url.OriginalURL{
				URL: "https://very-long-url.com",
			},
			want: "0",
		},
		{
			name:       "repeat url",
			filePath:   "test_data_file",
			fileExists: true,
			fileData:   `{"id":"0","url":"https://very-long-url-0.com"}` + "\n" + `{"id":"1","url":"https://very-long-url-1.com"}`,
			keeper:     NewKeeper("test_data_file"),
			url: url.OriginalURL{
				URL: "https://very-long-url-1.com",
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

			got, err := tt.keeper.Add(tt.url)

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
		keeper   url.DataKeeper
		id       string
		want     *url.OriginalURL
		wantErr  bool
	}{
		{
			name:     "ok",
			filePath: "test_data_file",
			fileData: `{"id":"0","url":"https://very-long-url.com"}`,
			keeper:   NewKeeper("test_data_file"),
			id:       "0",
			want: &url.OriginalURL{
				ID:  "0",
				URL: "https://very-long-url.com",
			},
			wantErr: false,
		},
		{
			name:     "not found",
			filePath: "test_data_file",
			fileData: `{"id":"0","url":"https://very-long-url-0.com"}` + "\n" + `{"id":"1","url":"https://very-long-url-1.com"}`,
			keeper:   NewKeeper("test_data_file"),
			id:       "2",
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := os.Create(tt.filePath)
			file.Write([]byte(tt.fileData))
			file.Close()

			got, err := tt.keeper.Get(tt.id)

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
