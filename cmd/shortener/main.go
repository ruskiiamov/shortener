package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	AuthSignKey     string `env:"AUTH_SIGN_KEY" envDefault:"secret_key"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func getConfig() *Config {
	var config Config

	env.Parse(&config)

	flag.StringVar(&config.ServerAddress, "a", config.ServerAddress, "Server address")
	flag.StringVar(&config.BaseURL, "b", config.BaseURL, "Base URL")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "File storage path")
	flag.StringVar(&config.AuthSignKey, "s", config.AuthSignKey, "Auth sign key")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "Database DSN")
	flag.Parse()

	return &config
}

func main() {
	config := getConfig()

	dataKeeper, err := data.NewKeeper(config.DatabaseDSN, config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer dataKeeper.Close()

	urlConverter := url.NewConverter(dataKeeper)

	router := chi.NewRouter()
	serverConfig := server.Config{
		BaseURL: config.BaseURL,
		SignKey: config.AuthSignKey,
	}
	handler := server.NewHandler(urlConverter, router, serverConfig)
	defer handler.Close()

	go func() {
		log.Fatal(http.ListenAndServe(config.ServerAddress, handler))
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
}
