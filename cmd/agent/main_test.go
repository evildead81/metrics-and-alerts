package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/agent"
	"github.com/evildead81/metrics-and-alerts/internal/server/handlers"
	memstorage "github.com/evildead81/metrics-and-alerts/internal/server/storages/mem-storage"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	storage := memstorage.New("./metrics.json", true)
	h := http.HandlerFunc(handlers.UpdateMetricByParamsHandler(storage))
	s := httptest.NewServer(h)
	defer s.Close()
	_, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	agent := agent.New("localhost:8080", 2*time.Second, 10*time.Second, ctx)
	err := agent.Run()
	require.NoError(t, err)
}
