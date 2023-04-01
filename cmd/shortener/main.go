// URL shortener service
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/config"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
)

const maxShutdownTime = 3 * time.Second

var (
	buildVersion string = `"N/A"`
	buildDate    string = `"N/A"`
	buildCommit  string = `"N/A"`
)

func main() {
	go func() {
		http.ListenAndServe(":9090", nil)
	}()

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config := config.Load()

	dataKeeper, err := data.NewKeeper(config.DatabaseDSN, config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	urlConverter := url.NewConverter(dataKeeper)
	router := chi.NewRouter()
	handler := server.NewHandler(ctx, urlConverter, router, config.BaseURL, config.AuthSignKey)

	manager := &autocert.Manager{Prompt: autocert.AcceptTOS}

	s := &http.Server{
		Addr:    config.ServerAddress,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		TLSConfig: manager.TLSConfig(),
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if config.EnableHTTPS {
			return s.ListenAndServeTLS("", "")
		}
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
