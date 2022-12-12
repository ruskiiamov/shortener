package main

import (
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
)

const (
	serverAddress = ":8080"
	baseURL       = "http://localhost:8080"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

func main() {
	config := Config{}
	env.Parse(&config)

	log.Println(config)

	dataKeeper := data.NewKeeper()
	urlConverter := url.NewConverter(dataKeeper, baseURL)

	router := chi.NewRouter()
	handler := server.NewHandler(urlConverter, router)

	log.Fatal(http.ListenAndServe(serverAddress, handler))
}
