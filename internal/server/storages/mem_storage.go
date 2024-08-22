package storages

import (
	"errors"
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
}

func (t *MemStorage) UpdateGauge(name string, value float64) {
	t.gaugeMetrics[name] = value
}

func (t MemStorage) GetCounters() map[string]int64 {
	return t.counterMetrics
}

func (t MemStorage) GetGauges() map[string]float64 {
	return t.gaugeMetrics
}

func (t MemStorage) GetGaugeValueByName(name string) (float64, error) {
	value, ok := t.gaugeMetrics[name]
	if !ok {
		return 0, errors.New("Gauge metric with name " + name + " not found")
	}
	return value, nil
}

func (t MemStorage) GetCountValueByName(name string) (int64, error) {
	value, ok := t.counterMetrics[name]
	if !ok {
		return 0, errors.New("Counter metric with name " + name + " not found")
	}
	return value, nil
}

func (t MemStorage) printCounters() {
	for key, value := range t.counterMetrics {
		fmt.Println("Counter", "Name", key, "Value", value)
	}
}
func (t MemStorage) printGauges() {
	for key, value := range t.gaugeMetrics {
		fmt.Println("Gauge", "Name", key, "Value", value)
	}
}

func New() *MemStorage {
	return &MemStorage{gaugeMetrics: make(map[string]float64), counterMetrics: map[string]int64{}}
}
