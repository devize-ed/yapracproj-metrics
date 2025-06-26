package main

import (
	"flag"
)

var (
	host string
)

func parseFlags() {
	flag.StringVar(&host, "a", ":8080", "address and port of the server")
	flag.Parse()
}
