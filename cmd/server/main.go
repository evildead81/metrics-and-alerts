package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/evildead81/metrics-and-alerts/internal/server"
	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
)

func main() {
	var endpointParam = flag.String("a", "localhost:8080", "server endpoint")
	flag.Parse()
	var cfg server.ServerConfig
	err := env.Parse(&cfg)

	var endpoint *string
	switch {
	case err == nil:
		if len(cfg.Address) != 0 {
			endpoint = &cfg.Address
		} else {
			endpoint = endpointParam
		}
	default:
		log.Fatal("Server env params parse error")
		endpoint = endpointParam
	}

	instance.New(*endpoint).Run()
}
