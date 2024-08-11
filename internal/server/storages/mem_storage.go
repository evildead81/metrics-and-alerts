package storages

import (
	"fmt"
)

type MemStorage struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	Storage
}

func (t *MemStorage) UpdateCounter(name string, value int64) {
	_, ok := t.counterMetrics[name]
	if ok {
		t.counterMetrics[name] += value
	} else {
		t.counterMetrics[name] = value
	}
	t.PrintValues()
}

func (t *MemStorage) UpdateGauge(name string, value float64) {
	t.gaugeMetrics[name] = value
	t.PrintValues()
}

func (t MemStorage) PrintValues() {
	for key, value := range t.counterMetrics {
		fmt.Println("Metric type: counter", "Name", key, "Val", value)
	}
	for key, value := range t.gaugeMetrics {
		fmt.Println("Metric type: gauge", "Name", key, "Val", value)
	}
}

func New() *MemStorage {
	return &MemStorage{gaugeMetrics: make(map[string]float64), counterMetrics: map[string]int64{}}
}
