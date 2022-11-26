package usecase

import "github.com/ruskiiamov/shortener/internal/entity"

type shortenerUseCase struct {
	repo ShortenerRepo
}

func NewShortenerUseCase(r ShortenerRepo) Shortener {
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
