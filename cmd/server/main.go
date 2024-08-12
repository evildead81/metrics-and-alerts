package main

import (
	"flag"

	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
)

func main() {
	enpoint := flag.String("a", "localhost:8080", "server endpoint")
	flag.Parse()
	instance.New(*enpoint).Run()
}
