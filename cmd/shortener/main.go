package main

import (
	"log"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/handler"
	"github.com/ruskiiamov/shortener/internal/usecase"
	"github.com/ruskiiamov/shortener/internal/usecase/repo"
)

func main() {
	shortenerRepo := repo.NewShortenerSlice()
	shortenerUseCase := usecase.NewShortener(shortenerRepo)
	handler := handler.New(shortenerUseCase)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
