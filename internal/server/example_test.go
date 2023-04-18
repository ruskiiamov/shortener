package server_test

import (
	"context"
	"log"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/chi"
	"github.com/ruskiiamov/shortener/internal/data"
	"github.com/ruskiiamov/shortener/internal/server"
	"github.com/ruskiiamov/shortener/internal/url"
	"github.com/ruskiiamov/shortener/internal/user"
)

func Example() {
	dataKeeper, err := data.NewKeeper("", "")
	if err != nil {
		log.Fatal(err)
	}

	userAuthorizer := user.NewAuthorizer([]byte("secret"))
	urlConverter := url.NewConverter(dataKeeper)
	delBuf := url.StartDeleteURL(context.Background(), urlConverter)

	router := chi.NewRouter()

	handler, err := server.NewHandler(
		context.Background(),
		userAuthorizer,
		urlConverter,
		router,
		delBuf,
		"http://localhost:8080",
		"",
	)
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	server.ListenAndServe()
}
