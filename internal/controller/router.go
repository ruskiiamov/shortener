package controller

import (
	"net/http"

	"github.com/ruskiiamov/shortener/internal/usecase"
)

func NewRouter(s usecase.Shortener) {
	http.HandleFunc("/", newShortenerHandler(s))
}
