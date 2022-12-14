package data

import (
	"errors"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type dataMemKeeper []string

func (d *dataMemKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	if id, ok := d.getID(originalURL.URL); ok {
		return strconv.Itoa(id), nil
	}

	id = strconv.Itoa(len(*d))
	*d = append(*d, originalURL.URL)

	return id, nil
}

func (d *dataMemKeeper) Get(id string) (*url.OriginalURL, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.New("wrong id")
	}

	if intID < 0 || intID >= len(*d) {
		return nil, errors.New("wrong id")
	}

	originalURL := &url.OriginalURL{
		ID:  id,
		URL: (*d)[intID],
	}

	return originalURL, nil
}

func (d *dataMemKeeper) getID(url string) (int, bool) {
	for id, originalURL := range *d {
		if originalURL == url {
			return id, true
		}
	}

	return 0, false
}
