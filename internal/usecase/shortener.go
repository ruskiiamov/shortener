package usecase

import (
	"github.com/ruskiiamov/shortener/internal/entity"
)

type ShortenerRepo interface {
	Add(entity.ShortenedURL) (id string, err error)
	Get(id string) (*entity.ShortenedURL, error)
}

type shortenerUseCase struct {
	repo ShortenerRepo
}

func NewShortener(r ShortenerRepo) *shortenerUseCase {
	return &shortenerUseCase{repo: r}
}

func (uc *shortenerUseCase) Shorten(host, url string) (string, error) {
	shortenedURL := entity.ShortenedURL{OriginalURL: url}

	id, err := uc.repo.Add(shortenedURL)
	if err != nil {
		return "", err
	}

	return host + "/" + id, nil
}

func (uc *shortenerUseCase) GetOriginal(id string) (string, error) {
	shortenedURL, err := uc.repo.Get(id)
	if err != nil {
		return "", err
	}

	return shortenedURL.OriginalURL, nil
}
