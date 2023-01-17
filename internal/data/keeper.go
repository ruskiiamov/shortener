package data

import "github.com/ruskiiamov/shortener/internal/url"

func NewKeeper(databaseDSN, filePath string) (url.DataKeeper, error) {
	if databaseDSN != "" {
		return newDBKeeper(databaseDSN)
	}

	return newMemKeeper(filePath)
}
