package main

import (
	"log"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/router"
	"github.com/ruskiiamov/shortener/internal/storage"
	"github.com/ruskiiamov/shortener/internal/url"
)

const port = ":8080"

func main() {
	urlStorage := storage.NewURLStorage()
	urlHandler := url.NewHandler(urlStorage)
	router := router.NewRouter(urlHandler)

	log.Fatal(http.ListenAndServe(port, router))
}
