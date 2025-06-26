package main

import (
	"flag"
	"time"
)

var (
	host           string
	popollInterval time.Duration
	reportInterval time.Duration
)

func parseFlags() {
	flag.StringVar(&host, "a", ":8080", "address and port of the server")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "reporting interval")
	flag.DurationVar(&popollInterval, "p", 2*time.Second, "polling interval")
	flag.Parse()
}
