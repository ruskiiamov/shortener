// URL shortener service
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/config"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/grpcserver"
	pb "github.com/ruskiiamov/shortener/internal/proto"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/ruskiiamov/shortener/internal/user"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	config, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	dataKeeper, err := data.NewKeeper(config.DatabaseDSN, config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}

	userAuthorizer := user.NewAuthorizer([]byte(config.AuthSignKey))
	urlConverter := url.NewConverter(dataKeeper)
	delBuf := url.StartDeleteURL(ctx, urlConverter)

	router := chi.NewRouter()
	handler, err := server.NewHandler(ctx, userAuthorizer, urlConverter, router, delBuf, config.BaseURL, config.TrustedSubnet)
	if err != nil {
		log.Fatal(err)
	}

	manager := &autocert.Manager{Prompt: autocert.AcceptTOS}

	httpServer := &http.Server{
		Addr:    config.ServerAddress,
		Handler: handler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		TLSConfig: manager.TLSConfig(),
	}

	listen, err := net.Listen("tcp", "127.0.0.1:3200")
	if err != nil {
		log.Fatal(err)
	}

	authInterceptor := grpcserver.NewAuthInterceptor(userAuthorizer)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))
	pb.RegisterShortenerServer(grpcServer, grpcserver.NewGRPCServer(urlConverter, delBuf))

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if config.EnableHTTPS {
			return httpServer.ListenAndServeTLS("", "")
		}
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		return grpcServer.Serve(listen)
	})

	g.Go(func() error {
		<-gCtx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		grpcServer.Stop()

		err = httpServer.Shutdown(ctx)
		if err != nil {
			log.Printf("server shutdown error: %s", err)
		}

		close(delBuf)

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
