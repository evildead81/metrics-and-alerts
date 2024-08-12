package storages

type Storage interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	GetCounters() map[string]int64
	GetGauges() map[string]float64
	GetGaugeValueByName(name string) (float64, error)
	GetCountValueByName(name string) (int64, error)
}
