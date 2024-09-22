package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/evildead81/metrics-and-alerts/internal/server"
	"github.com/evildead81/metrics-and-alerts/internal/server/instance"
	"github.com/evildead81/metrics-and-alerts/internal/server/logger"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
	dbstorage "github.com/evildead81/metrics-and-alerts/internal/server/storages/db-storage"
	memstorage "github.com/evildead81/metrics-and-alerts/internal/server/storages/mem-storage"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func tryOpenDB(connStr string, attempt uint8) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		fmt.Println(err)
		if errors.As(err, pgerrcode.ConnectionException) {
			if attempt < 3 {
				attempt := attempt + 1
				time.Sleep(time.Duration(attempt*2-1) * time.Second)
				tryOpenDB(connStr, attempt+1)
			}
		}
		return nil, err
	}
	return db, nil
}

func main() {
	var endpointParam = flag.String("a", "localhost:8080", "Server endpoint")
	var storeIntervalParam = flag.Int64("i", 300, "Save metrics into file interval")
	var fileStoragePathParam = flag.String("f", "./metrics.json", "File storage path")
	var restoreParam = flag.Bool("r", true, "Restore from file flag")
	var connStrParam = flag.String("d", "", "DB connection string")
	flag.Parse()
	var cfg server.ServerConfig
	err := env.Parse(&cfg)

	var endpoint *string
	var storeInterval *int64
	var fileStoragePath *string
	var restore *bool
	var connStr *string
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
		if cfg.Restore {
			restore = &cfg.Restore
		} else {
			restore = restoreParam
		}
		if len(cfg.DatabaseDSN) != 0 {
			connStr = &cfg.DatabaseDSN
		} else {
			connStr = connStrParam
		}
	default:
		logger.Logger.Fatalw("Server env params parse error", "error", err.Error())
		endpoint = endpointParam
	}

	var storage storages.Storage
	if len(*connStr) != 0 {
		db, err := tryOpenDB(*connStr, 0)
		if err != nil {
			logger.Logger.Fatalw("Error while connect to DB", "error", err.Error())
		}
		defer db.Close()
		storage = dbstorage.New(db)
	} else {
		storage = memstorage.New(*fileStoragePath, *restore)
	}

	instance.New(
		*endpoint,
		&storage,
		time.Duration(*storeInterval)*time.Second,
	).Run()
}
