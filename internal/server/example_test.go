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

	handler := server.NewHandler(context.Background(), urlConverter, router, "http://localhost:8080", "secret_key")

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	server.ListenAndServe()
}
