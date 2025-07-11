package main

import (
	"fmt"

	"log"

	"github.com/devize-ed/yapracproj-metrics.git/internal/config"
	"github.com/devize-ed/yapracproj-metrics.git/internal/handler"
	st "github.com/devize-ed/yapracproj-metrics.git/internal/repository"
	"github.com/devize-ed/yapracproj-metrics.git/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	ms := st.NewMemStorage() // init the memory storage for metrics
	h := handler.NewHandler(ms)
	cfg := config.GetServerConfig() // call the function to parse cl flags

	srv := server.NewServer(cfg, h) // create a new server with the configuration and handler

	// loging the address and starting the server
	log.Println("Starting HTTP server on ", cfg.Host)
	err := srv.ListenAndServe()

	return fmt.Errorf("failed to start HTTP server: %w", err)
}
