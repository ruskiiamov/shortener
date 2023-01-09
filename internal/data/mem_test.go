package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var keeper *memKeeper

func init() {
	urls := map[int]memURL{
		1: {
			Original: "http://shortener.com",
			Users:    []string{"c7cbe16d-034e-40b9-a2a5-e936851c4282"},
		},
		2: {
			Original: "http://shortener.com/info",
			Users:    []string{"b01ad148-d4da-4b08-9c75-9eb66899119f"},
		},
		3: {
			Original: "http://shortener.com/stat",
			Users:    []string{"b01ad148-d4da-4b08-9c75-9eb66899119f", "c7cbe16d-034e-40b9-a2a5-e936851c4282"},
		},
	}

	data := URLData{
		NextID: 4,
		URLs:   urls,
	}

	keeper = &memKeeper{data: data}
}

func TestMemAdd(t *testing.T) {
	tests := []struct {
		name     string
		original string
		userID   string
		wantErr  bool
	}{
		{
			name:     "ok",
			original: "http://shortener.com/other",
			userID:   "1770aae6-caaf-4578-b27e-ffa967927a1b",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := keeper.Add(tt.userID, tt.original)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, id)
		})
	}
}

func TestMemGet(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		want    string
		wantErr bool
	}{
		{
			name:    "ok",
			id:      1,
			want:    "http://shortener.com",
			wantErr: false,
		},
		{
			name:    "wrong id",
			id:      0,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keeper.Get(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAllByUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		want    map[string]int
		wantErr bool
	}{
		{
			name:    "ok",
			userID:  "b01ad148-d4da-4b08-9c75-9eb66899119f",
			want:    map[string]int{"http://shortener.com/info": 2, "http://shortener.com/stat": 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keeper.GetAllByUser(tt.userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
