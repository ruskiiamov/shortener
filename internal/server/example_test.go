package server_test

import (
	"context"
	"log"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
)

func Example() {
	dataKeeper, err := data.NewKeeper("", "")
	if err != nil {
		log.Fatal(err)
	}

	urlConverter := url.NewConverter(dataKeeper)

	router := chi.NewRouter()

	serverConfig := server.Config{
		BaseURL: "http://localhost:8080",
		SignKey: "secret_key",
	}

	handler := server.NewHandler(context.Background(), urlConverter, router, serverConfig)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	server.ListenAndServe()
}
