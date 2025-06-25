package main

import (
	"fmt"
	"net/http"

	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	"github.com/devize-ed/yapracproj-metrics.git/internal/repository/storage"
)

const host = ":8080"

func main() {
	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	storage := storage.NewMemStorage()
	mux := http.NewServeMux()
	mux.Handle(`/update/`, handler.Middleware(handler.UpdateHandler(storage)))

	fmt.Println("Starting HTTP server on ", host)
	err := http.ListenAndServe(host, mux)
	return err
}
