package storages

import "fmt"

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
	t.printCounters()
}

func (t *MemStorage) UpdateGauge(name string, value float64) {
	t.gaugeMetrics[name] = value
	t.printGauges()
}

func (t MemStorage) GetCounters() map[string]int64 {
	return t.counterMetrics
}

func (t MemStorage) GetGauges() map[string]float64 {
	return t.gaugeMetrics
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
