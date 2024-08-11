package storages

type Storage interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
}
