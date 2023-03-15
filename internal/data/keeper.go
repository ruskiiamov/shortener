// Package data is the data storage abstraction for URLs.
package data

import "github.com/ruskiiamov/shortener/internal/url"

// NewKeeper returns object that implements url.DataKeeper interface.
//
// If databaseDSN provided, NewKeeper returns DB implementation.
// Otherwise NewKeeper returns in-memory implementation with dumps
// to filePath.
func NewKeeper(databaseDSN, filePath string) (url.DataKeeper, error) {
	if databaseDSN != "" {
		return newDBKeeper(databaseDSN)
	}

	return newMemKeeper(filePath)
}
