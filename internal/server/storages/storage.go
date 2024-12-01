package storages

import "github.com/evildead81/metrics-and-alerts/internal/contracts"

// Storage - интерфейс хранилища, с которым работают обработчики запросов.
type Storage interface {
	// UpdateCounter обновляет метрику типа Counter.
	UpdateCounter(name string, value int64) error
	// UpdateGauge обновляет метрику типа Gauge.
	UpdateGauge(name string, value float64) error
	// GetCounters возвращает метрики типа Counter.
	GetCounters() map[string]int64
	// GetGauges возвращает метрики типа Gauge.
	GetGauges() map[string]float64
	// GetGaugeValueByName возвращает метрику типа Gauge по имени.
	GetGaugeValueByName(name string) (float64, error)
	// GetCountValueByName возвращает метрику типа Counter по имени.
	GetCountValueByName(name string) (int64, error)
	// Restore восстанавливает хранилище при запуске сервера.
	Restore() error
	// Write сохраняет данные в хранилище.
	Write() error
	// Ping проверяет доступность хранилища.
	Ping() error
	// UpdateMetrics обновляет список метрик.
	UpdateMetrics(metrics []contracts.Metrics) error
}
