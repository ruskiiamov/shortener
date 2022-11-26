package usecase

import "github.com/ruskiiamov/shortener/internal/entity"

type (
	Shortener interface {
		Shorten(host, url string) (string, error)
		GetOriginal(id string) (string, error)
	}

	ShortenerRepo interface {
		Add(entity.ShortenedURL) (id string, err error)
		Get(id string) (*entity.ShortenedURL, error)
	}
)
