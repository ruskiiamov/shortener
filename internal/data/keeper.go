package data

import "github.com/ruskiiamov/shortener/internal/url"

func NewKeeper(databaseDSN, filePath string) (url.DataKeeper, error) {
	switch {
	case databaseDSN != "":
		return newDBKeeper(databaseDSN)
	case filePath != "":
		return newFileKeeper(filePath), nil
	default:
		return newMemKeeper(), nil
	}
}
