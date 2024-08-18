package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/evildead81/metrics-and-alerts/internal/agent"
)

func main() {
	var endpointParam = flag.String("a", "localhost:8080", "server endpoint")
	var reportIntervalParam = flag.Int64("r", 10, "Report interval")
	var pollIntervalParam = flag.Int64("p", 2, "Poll interval")
	flag.Parse()
	var cfg agent.AgentConfig
	err := env.Parse(&cfg)

	var endpoint *string
	var reportInterval *int64
	var pollInterval *int64
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
		}
	default:
		log.Fatal("Agent env params parse error")
		endpoint = endpointParam
		reportInterval = reportIntervalParam
		pollInterval = pollIntervalParam
	}

	fmt.Println(*endpoint)

	agent.New(
		*endpoint,
		time.Duration(*pollInterval)*time.Second,
		time.Duration(*reportInterval)*time.Second,
		context.Background(),
	).Run()
}
