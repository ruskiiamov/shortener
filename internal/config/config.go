// Package config is the configuration provider for the whole application.
package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config for env parsing.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080" json:"server_address"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	AuthSignKey     string `env:"AUTH_SIGN_KEY" envDefault:"secret_key" json:"auth_sign_key"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	Config          string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
}

// Load returns structure with configuration parameters.
func Load() (*Config, error) {
	var config Config

	env.Parse(&config)

	flag.StringVar(&config.ServerAddress, "a", config.ServerAddress, "Server address")
	flag.StringVar(&config.BaseURL, "b", config.BaseURL, "Base URL")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "File storage path")
	flag.StringVar(&config.AuthSignKey, "k", config.AuthSignKey, "Auth sign key")
	flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "Database DSN")
	flag.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "Enables HTTPS")
	flag.StringVar(&config.Config, "config", config.Config, "Configuration file path")
	flag.StringVar(&config.Config, "c", config.Config, "Configuration file path (shorthand)")
	flag.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "")
	flag.Parse()

	if config.Config == "" {
		return &config, nil
	}

	file, err := os.OpenFile(config.Config, os.O_RDONLY, 0777)
	if err != nil {
		return nil, err
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var jsonConfig Config
	err = json.Unmarshal(fileData, &jsonConfig)
	if err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = jsonConfig.ServerAddress
	}

	if config.BaseURL == "" {
		config.BaseURL = jsonConfig.BaseURL
	}

	if config.FileStoragePath == "" {
		config.FileStoragePath = jsonConfig.FileStoragePath
	}

	if config.AuthSignKey == "" {
		config.AuthSignKey = jsonConfig.AuthSignKey
	}

	if config.DatabaseDSN == "" {
		config.DatabaseDSN = jsonConfig.DatabaseDSN
	}

	if !config.EnableHTTPS {
		config.EnableHTTPS = jsonConfig.EnableHTTPS
	}

	if config.TrustedSubnet == "" {
		config.TrustedSubnet = jsonConfig.TrustedSubnet
	}

	return &config, nil
}
