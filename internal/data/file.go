package data

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"github.com/ruskiiamov/shortener/internal/url"
)

type fileKeeper struct {
	filePath string
}

func newFileKeeper(filePath string) url.DataKeeper {
	return &fileKeeper{filePath: filePath}
}

func (f *fileKeeper) Add(originalURL url.OriginalURL) (id string, err error) {
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		file.Close()
		return "", err
	}
	defer file.Close()

	id, count, ok := f.getID(originalURL, file)
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

func (f *fileKeeper) Get(id string) (*url.OriginalURL, error) {
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_RDONLY, 0777)
	if err != nil {
		file.Close()
		return nil, err
	}
	defer file.Close()

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

func (f *fileKeeper) GetAllByUser(userID string) ([]url.OriginalURL, error) {
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_RDONLY, 0777)
	if err != nil {
		file.Close()
		return nil, err
	}
	defer file.Close()

	fileDec := json.NewDecoder(file)

	res := make([]url.OriginalURL, 0)
	var url url.OriginalURL

	for fileDec.Decode(&url) == nil {
		if url.UserID == userID {
			res = append(res, url)
		}
	}

	return res, nil
}

func (f *fileKeeper) PingDB() error {
	return errors.New("file data keeper is used")
}

func (f *fileKeeper) Close() {}

func (f *fileKeeper) getID(originalURL url.OriginalURL, file *os.File) (string, int, bool) {
	fileDec := json.NewDecoder(file)

	var url url.OriginalURL
	count := 0

	for {
		err := fileDec.Decode(&url)
		if err != nil {
			return "", count, false
		}

		if url.URL == originalURL.URL && url.UserID == originalURL.UserID {
			return url.ID, count, true
		}

		count++
	}
}
