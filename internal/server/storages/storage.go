package storages

import "github.com/evildead81/metrics-and-alerts/internal/contracts"

type Storage interface {
	UpdateCounter(name string, value int64) error
	UpdateGauge(name string, value float64) error
	GetCounters() map[string]int64
	GetGauges() map[string]float64
	GetGaugeValueByName(name string) (float64, error)
	GetCountValueByName(name string) (int64, error)
	Restore() error
	Write() error
	Ping() error
	UpdateMetrics(metrics []contracts.Metrics) error
}
