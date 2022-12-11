package main

import (
	"log"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
)

const port = ":8080"

func main() {
	dataKeeper := data.NewKeeper()
	urlConverter := url.NewConverter(dataKeeper)

	router := chi.NewRouter()
	handler := server.NewHandler(urlConverter, router)

	log.Fatal(http.ListenAndServe(port, handler))
}
