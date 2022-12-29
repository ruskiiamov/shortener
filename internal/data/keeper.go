package data

import "github.com/ruskiiamov/shortener/internal/url"

func NewKeeper(filePath string) url.DataKeeper {
	if filePath == "" {
		return new(dataMemKeeper)
	}

	return &dataFileKeeper{filePath: filePath}
}
