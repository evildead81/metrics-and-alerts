package main

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
	"github.com/evildead81/metrics-and-alerts/internal/server"
	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
)

func main() {
	var endpointParam = flag.String("a", "localhost:8080", "Server endpoint")
	var storeIntervalParam = flag.Int64("i", 300, "Save metrics into file interval")
	var fileStoragePathParam = flag.String("f", "./metrics.json", "File storage path")
	var restoreParam = flag.Bool("r", true, "Restore from file flag")
	flag.Parse()
	var cfg server.ServerConfig
	err := env.Parse(&cfg)

	var endpoint *string
	var storeInterval *int64
	var fileStoragePath *string
	var restore *bool
	switch {
	case err == nil:
		if len(cfg.Address) != 0 {
			endpoint = &cfg.Address
		} else {
			endpoint = endpointParam
		}
		if cfg.StoreInterval != 0 {
			storeInterval = &cfg.StoreInterval
		} else {
			storeInterval = storeIntervalParam
		}
		if len(cfg.FileStoragePath) != 0 {
			fileStoragePath = &cfg.FileStoragePath
		} else {
			fileStoragePath = fileStoragePathParam
		}
		if cfg.Restore != false {
			restore = &cfg.Restore
		} else {
			restore = restoreParam
		}
	default:
		logger.Logger.Fatalw("Server env params parse error", "error", err.Error())
		endpoint = endpointParam
	}

	instance.New(
		*endpoint,
		time.Duration(*storeInterval)*time.Second,
		*fileStoragePath,
		*restore,
	).Run()
}
