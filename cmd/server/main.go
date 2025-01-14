package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/evildead81/metrics-and-alerts/internal/server/config"
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
		if pgerrcode.IsConnectionException(err.Error()) {
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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildParams() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	var endpointParam = flag.String("a", "localhost:8080", "Server endpoint")
	var storeIntervalParam = flag.Int64("i", 300, "Save metrics into file interval")
	var fileStoragePathParam = flag.String("f", "./metrics.json", "File storage path")
	var restoreParam = flag.Bool("r", true, "Restore from file flag")
	var connStrParam = flag.String("d", "", "DB connection string")
	var keyParam = flag.String("k", "", "Secret key")
	var cryptoKeyPathParam = flag.String("crypto-key", "", "Public key")
	var trustedSubnetParam = flag.String("t", "", "Trusted subnet")
	var useRPCParam = flag.Bool("use-rpc", false, "Use server as gRPC")
	var configPathParam = flag.String("c", "", "Config path")
	flag.Parse()
	var cfg config.ServerConfig
	err := env.Parse(&cfg)

	var endpoint *string
	var storeInterval *int64
	var fileStoragePath *string
	var restore *bool
	var connStr *string
	var key *string
	var cryptoKeyPath *string
	var trustedSubnet *string
	var configPath *string
	var useRPC *bool
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
		if len(cfg.Key) != 0 {
			key = &cfg.Key
		} else {
			key = keyParam
		}
		if cfg.CryptoKey != "" {
			cryptoKeyPath = &cfg.CryptoKey
		} else {
			cryptoKeyPath = cryptoKeyPathParam
		}
		if cfg.TrustedSubnet != "" {
			trustedSubnet = &cfg.TrustedSubnet
		} else {
			trustedSubnet = trustedSubnetParam
		}
		if !cfg.UseRPC {
			useRPC = &cfg.UseRPC
		} else {
			useRPC = useRPCParam
		}

		if cfg.ConfigPath != "" {
			configPath = &cfg.ConfigPath
		} else {
			configPath = configPathParam
		}
	default:
		logger.Logger.Fatalw("Server env params parse error", "error", err.Error())
		endpoint = endpointParam
	}

	if len(*configPath) != 0 {
		content, err := os.ReadFile(*configPath)
		if err != nil {
			panic(err)
		}

		var fConfig config.ServerConfig
		err = json.Unmarshal(content, &fConfig)
		if err != nil {
			panic(err)
		}

		if len(*endpoint) == 0 {
			endpoint = &fConfig.Address
		}
		if *storeInterval == 0 {
			storeInterval = &fConfig.StoreInterval
		}
		if len(*fileStoragePath) == 0 {
			fileStoragePath = &fConfig.FileStoragePath
		}
		if !*restore {
			restore = &fConfig.Restore
		}
		if len(*cryptoKeyPath) == 0 {
			cryptoKeyPath = &fConfig.CryptoKey
		}
		if len(*connStr) == 0 {
			connStr = &fConfig.DatabaseDSN
		}
		if len(*trustedSubnet) == 0 {
			trustedSubnet = &fConfig.TrustedSubnet
		}
		if len(*key) == 0 {
			key = &fConfig.Key
		}
		if !*useRPC {
			useRPC = &fConfig.UseRPC
		}
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

	printBuildParams()

	servInstance := instance.New(
		*endpoint,
		&storage,
		time.Duration(*storeInterval)*time.Second,
		*key,
		*cryptoKeyPath,
		*trustedSubnet,
	)

	if *useRPC {
		servInstance.RunRPC()
	} else {
		servInstance.Run()
	}
}
