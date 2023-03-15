// URL shortener service
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
	"golang.org/x/sync/errgroup"
)

const maxShutdownTime = 3 * time.Second

// Config for env parsing.
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
	go func() {
		http.ListenAndServe(":9090", nil)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config := getConfig()

	dataKeeper, err := data.NewKeeper(config.DatabaseDSN, config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	urlConverter := url.NewConverter(dataKeeper)

	router := chi.NewRouter()
	serverConfig := server.Config{
		BaseURL: config.BaseURL,
		SignKey: config.AuthSignKey,
	}
	handler := server.NewHandler(ctx, urlConverter, router, serverConfig)

	s := &http.Server{
		Addr:    config.ServerAddress,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = s.Shutdown(ctx)
		if err != nil {
			log.Printf("server shutdown error: %s", err)
		}

		err = handler.Close(ctx)
		if err != nil {
			log.Printf("handler close error: %s", err)
		}

		err = dataKeeper.Close(ctx)
		if err != nil {
			log.Printf("data keeper close error: %s", err)
		}

		return err
	})

	if err = g.Wait(); err != nil {
		log.Printf("exit: %s", err)
	}
}
