package data

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type dataFileKeeper struct {
	filePath string
}

func (d *dataFileKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	file, err := os.OpenFile(d.filePath, os.O_CREATE|os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		return "", err
	}

	id, count, ok := d.getID(originalURL.URL, file)
	if ok {
		return id, nil
	}

	originalURL.ID = strconv.Itoa(count)

	fileEnc := json.NewEncoder(file)

	err = fileEnc.Encode(originalURL)
	if err != nil {
		return "", err
	}

	return originalURL.ID, nil
}

func (d *dataFileKeeper) Get(id string) (*url.OriginalURL, error) {
	file, err := os.OpenFile(d.filePath, os.O_CREATE|os.O_RDONLY, 0777)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	fileDec := json.NewDecoder(file)

	var url url.OriginalURL

	for {
		err := fileDec.Decode(&url)
		if err != nil {
			return nil, errors.New("wrong id")
		}

		if url.ID == id && url.URL != "" {
			return &url, nil
		}
	}
}

func (d *dataFileKeeper) getID(originalURL string, file *os.File) (string, int, bool) {
	fileDec := json.NewDecoder(file)

	var url url.OriginalURL
	count := 0

	for {
		err := fileDec.Decode(&url)
		if err != nil {
			return "", count, false
		}

		if url.URL == originalURL {
			return url.ID, count, true
		}

		count++
	}
}
