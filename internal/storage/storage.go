package storage

import (
	"errors"
	neturl "net/url"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type storage []string

func NewURLStorage() *storage {
	return new(storage)
}

func (s *storage) Add(originalURL url.OriginalURL) (id string, err error) {
	if _, err := neturl.ParseRequestURI(originalURL.URL); err != nil {
		return "", err
	}

	if id, ok := s.getID(originalURL.URL); ok {
		return strconv.Itoa(id), nil
	}

	id = strconv.Itoa(len(*s))
	*s = append(*s, originalURL.URL)

	return id, nil
}

func (s *storage) Get(id string) (*url.OriginalURL, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("wrong id")
	}

	if intID < 0 || intID >= len(*s) {
		return nil, errors.New("wrong id")
	}

	originalURL := &url.OriginalURL{
		ID:  id,
		URL: (*s)[intID],
	}

	return originalURL, nil
}

func (s *storage) getID(url string) (int, bool) {
	for id, originalURL := range *s {
		if originalURL == url {
			return id, true
		}
	}

	return 0, false
}
