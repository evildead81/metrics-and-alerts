package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	_ "net/http/pprof"

	"github.com/caarlos0/env/v6"
	"github.com/evildead81/metrics-and-alerts/internal/agent"
)

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
	var reportIntervalParam = flag.Int64("r", 10, "Report interval")
	var pollIntervalParam = flag.Int64("p", 2, "Poll interval")
	var keyParam = flag.String("k", "", "Secret key")
	var rateLimitParam = flag.Int("l", 0, "Parallels sends cound")
	flag.Parse()
	var cfg agent.AgentConfig
	err := env.Parse(&cfg)

	var endpoint *string
	var reportInterval *int64
	var pollInterval *int64
	var key *string
	var rateLimit *int
	switch {
	case err == nil:
		{
			if len(cfg.Address) != 0 {
				endpoint = &cfg.Address
			} else {
				endpoint = endpointParam
			}
			if cfg.ReportInterval != 0 {
				reportInterval = &cfg.ReportInterval
			} else {
				reportInterval = reportIntervalParam
			}
			if cfg.PollInterval != 0 {
				pollInterval = &cfg.PollInterval
			} else {
				pollInterval = pollIntervalParam
			}
			if len(cfg.Key) != 0 {
				key = &cfg.Key
			} else {
				key = keyParam
			}
			if cfg.RateLimit != 0 {
				rateLimit = &cfg.RateLimit
			} else {
				rateLimit = rateLimitParam
			}
		}
	default:
		log.Fatal("Agent env params parse error")
		endpoint = endpointParam
		reportInterval = reportIntervalParam
		pollInterval = pollIntervalParam
	}

	printBuildParams()

	agent.New(
		*endpoint,
		time.Duration(*pollInterval)*time.Second,
		time.Duration(*reportInterval)*time.Second,
		context.Background(),
		*key,
		*rateLimit,
	).Run()
}
