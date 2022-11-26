package app

import (
	"github.com/ruskiiamov/shortener/internal/controller"
	"github.com/ruskiiamov/shortener/internal/usecase"
	"github.com/ruskiiamov/shortener/internal/usecase/repo"
	"github.com/ruskiiamov/shortener/pkg/server"
)

func Run() {
	shortenerRepo := repo.NewShortenerSlice()
	shortenerUseCase := usecase.NewShortenerUseCase(shortenerRepo)

	controller.NewRouter(shortenerUseCase)
	server.Run("8080")
}
