package repo

import (
	"errors"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/entity"
)

type shortenerSlice struct {
	s []string
}

func NewShortenerSlice() *shortenerSlice {
	return &shortenerSlice{
		s: []string{},
	}
}

func (s *shortenerSlice) Add(url entity.ShortenedURL) (id string, err error) {
	if id, ok := s.getID(url.OriginalURL); ok {
		return strconv.Itoa(id), nil
	}

	id = strconv.Itoa(len(s.s))
	s.s = append(s.s, url.OriginalURL)

	return id, nil
}

func (s *shortenerSlice) Get(id string) (*entity.ShortenedURL, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("wrong id")
	}

	if intID < 0 || intID >= len(s.s) {
		return nil, errors.New("wrong id")
	}

	shortenedURL := &entity.ShortenedURL{
		ID: id,
		OriginalURL: s.s[intID],
	}

	return shortenedURL, nil
}

func (s *shortenerSlice) getID(url string) (int, bool) {
	for id, originalURL := range s.s {
		if originalURL == url {
			return id, true
		}
	}

	return 0, false
}