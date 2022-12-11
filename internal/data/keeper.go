package data

import (
	"errors"
	neturl "net/url"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type dataKeeper []string

func NewKeeper() *dataKeeper {
	return new(dataKeeper)
}

func (d *dataKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	if _, err := neturl.ParseRequestURI(originalURL.URL); err != nil {
		return "", err
	}

	if id, ok := d.getID(originalURL.URL); ok {
		return strconv.Itoa(id), nil
	}

	id = strconv.Itoa(len(*d))
	*d = append(*d, originalURL.URL)

	return id, nil
}

func (d *dataKeeper) Get(id string) (*url.OriginalURL, error) {
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

func (d *dataKeeper) getID(url string) (int, bool) {
	for id, originalURL := range *d {
		if originalURL == url {
			return id, true
		}
	}

	return 0, false
}
