package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func getConfig() *Config {
	var config Config

	env.Parse(&config)

	flag.StringVar(&config.ServerAddress, "a", config.ServerAddress, "Server address")
	flag.StringVar(&config.BaseURL, "b", config.BaseURL, "Base URL")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "File storage path")
	flag.Parse()

	return &config
}

func main() {
	config := getConfig()

	dataKeeper := data.NewKeeper(config.FileStoragePath)
	urlConverter := url.NewConverter(dataKeeper, config.BaseURL)

	router := chi.NewRouter()
	handler := server.NewHandler(urlConverter, router)

	log.Fatal(http.ListenAndServe(config.ServerAddress, handler))
}
