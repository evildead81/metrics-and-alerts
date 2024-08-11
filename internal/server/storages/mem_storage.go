package storages

import (
	"fmt"
)

type MetricValues struct {
	gauge   float64
	counter int64
}
type MemStorage struct {
	metrics map[string]*MetricValues
	Storage
}

func (t *MemStorage) UpdateCounter(name string, value int64) {
	metric, ok := t.metrics[name]
	if ok {
		metric.counter += value
	} else {
		t.metrics[name] = &MetricValues{
			counter: value,
			gauge:   0,
		}
	}
	t.PrintValues()
}

func (t *MemStorage) UpdateGauge(name string, value float64) {
	metric, ok := t.metrics[name]
	if ok {
		metric.gauge = value
	} else {
		t.metrics[name] = &MetricValues{
			counter: 0,
			gauge:   value,
		}
	}
	t.PrintValues()
}

func (t MemStorage) PrintValues() {
	for key, value := range t.metrics {
		fmt.Println("Metric name", key, "gauge", value.gauge, "counter", value.counter)
	}
}

func New() *MemStorage {
	return &MemStorage{metrics: make(map[string]*MetricValues)}
}
