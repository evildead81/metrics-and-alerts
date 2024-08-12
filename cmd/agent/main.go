package main

import (
	"context"
	"flag"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/agent"
)

func main() {
	enpoint := flag.String("a", "http://localhost:8080", "server endpoint")
	reportInterval := flag.Int("r", 10, "Report interval")
	pollInterval := flag.Int("p", 2, "Poll interval")
	flag.Parse()
	ctx := context.Background()
	agent.New(*enpoint, time.Duration(*pollInterval)*time.Second, time.Duration(*reportInterval)*time.Second, ctx).Run()
}
