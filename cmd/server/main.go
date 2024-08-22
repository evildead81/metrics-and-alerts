package main

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/evildead81/metrics-and-alerts/internal/server"
	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
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
		logger.Logger.Fatalw("Server env params parse error", "error", err.Error())
		endpoint = endpointParam
	}

	instance.New(*endpoint).Run()
}
