package main

import (
	"context"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/agent"
)

func main() {
	ctx := context.Background()
	agent.New("http://localhost:8080", 2*time.Second, 10*time.Second, ctx).Run()
}
