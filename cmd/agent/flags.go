package main

import (
	"flag"
)

var (
	host           string
	pollInterval   int
	reportInterval int
)

func parseFlags() {
	flag.StringVar(&host, "a", ":8080", "address and port of the server")
	flag.IntVar(&reportInterval, "r", 10, "reporting interval in seconds")
	flag.IntVar(&pollInterval, "p", 2, "polling interval in seconds")
	flag.Parse()
}
