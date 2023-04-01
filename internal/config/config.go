// Package config is the configuration provider for the whole application.
package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

// Config for env parsing.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	AuthSignKey     string `env:"AUTH_SIGN_KEY" envDefault:"secret_key"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
}

// Load returns structure with configuration parameters.
func Load() *Config {
	var config Config

	env.Parse(&config)

	flag.StringVar(&config.ServerAddress, "a", config.ServerAddress, "Server address")
	flag.StringVar(&config.BaseURL, "b", config.BaseURL, "Base URL")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "File storage path")
	flag.StringVar(&config.AuthSignKey, "k", config.AuthSignKey, "Auth sign key")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "Database DSN")
	flag.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "Enables HTTPS")
	flag.Parse()

	return &config
}
