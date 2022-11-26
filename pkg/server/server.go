package server

import (
	"log"
	"net/http"
)

func Run(port string) {
	if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
		log.Fatalf("server error: %s", err.Error())
	}
}
