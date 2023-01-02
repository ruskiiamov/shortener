package data

import (
	"errors"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type dataMemKeeper []url.OriginalURL

func (d *dataMemKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	if id, ok := d.getID(originalURL); ok {
		return strconv.Itoa(id), nil
	}

	originalURL.ID = strconv.Itoa(len(*d))
	*d = append(*d, originalURL)

	return originalURL.ID, nil
}

func (d *dataMemKeeper) Get(id string) (*url.OriginalURL, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("wrong id")
	}

	if intID < 0 || intID >= len(*d) {
		return nil, errors.New("wrong id")
	}

	originalURL := &(*d)[intID]

	return originalURL, nil
}

func (d *dataMemKeeper) GetAllByUser(userID string) ([]url.OriginalURL, error) {
	var res []url.OriginalURL

	for _, originalURL := range *d {
		if originalURL.UserID == userID {
			res = append(res, originalURL)
		}
	}

	return res, nil
}

func (d *dataMemKeeper) getID(url url.OriginalURL) (int, bool) {
	for id, originalURL := range *d {
		if originalURL.URL == url.URL && originalURL.UserID == url.UserID {
			return id, true
		}
	}

	return 0, false
}
