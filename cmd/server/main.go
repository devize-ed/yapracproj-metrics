package main

import (
	"fmt"
	"net/http"
)

const host = ":8080"

func main() {
	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.Handle(`/update/`, middleware(makeUpdateHandler(storage)))

	fmt.Println("Starting HTTP server on ", host)
	err := http.ListenAndServe(host, mux)
	return err
}
