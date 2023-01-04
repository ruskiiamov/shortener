package data

import (
	"errors"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type memKeeper []url.OriginalURL

func newMemKeeper() url.DataKeeper {
	return new(memKeeper)
}

func (m *memKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	if id, ok := m.getID(originalURL); ok {
		return strconv.Itoa(id), nil
	}

	originalURL.ID = strconv.Itoa(len(*m))
	*m = append(*m, originalURL)

	return originalURL.ID, nil
}

func (m *memKeeper) Get(id string) (*url.OriginalURL, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("wrong id")
	}

	if intID < 0 || intID >= len(*m) {
		return nil, errors.New("wrong id")
	}

	originalURL := &(*m)[intID]

	return originalURL, nil
}

func (m *memKeeper) GetAllByUser(userID string) ([]url.OriginalURL, error) {
	res := make([]url.OriginalURL, 0)

	for _, originalURL := range *m {
		if originalURL.UserID == userID {
			res = append(res, originalURL)
		}
	}

	return res, nil
}

func (m *memKeeper) PingDB() error {
	return errors.New("memory data keeper is used")
}

func (m *memKeeper) Close() {}

func (m *memKeeper) getID(url url.OriginalURL) (int, bool) {
	for id, originalURL := range *m {
		if originalURL.URL == url.URL && originalURL.UserID == url.UserID {
			return id, true
		}
	}

	return 0, false
}
