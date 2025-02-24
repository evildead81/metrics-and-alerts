package memstorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

type MemStorage struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	storagePath    string
	mutex          *sync.Mutex
	storages.Storage
}

func New(storagePath string, restore bool) *MemStorage {
	storage := &MemStorage{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
		storagePath:    storagePath,
		mutex:          &sync.Mutex{},
	}

	if restore {
		storage.Restore()
	}

	return storage
}

func (t *MemStorage) UpdateCounter(name string, value int64) error {
	_, ok := t.counterMetrics[name]
	t.mutex.Lock()
	if ok {
		t.counterMetrics[name] += value
	} else {
		t.counterMetrics[name] = value
	}
	t.mutex.Unlock()

	return nil
}

func (t *MemStorage) UpdateGauge(name string, value float64) error {
	t.mutex.Lock()
	t.gaugeMetrics[name] = value
	t.mutex.Unlock()
	return nil
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

func (t MemStorage) Restore() error {
	content, err := os.ReadFile(t.storagePath)
	if err != nil {
		return err
	}

	var metrics []contracts.Metrics
	err = json.Unmarshal(content, &metrics)

	if err != nil {
		return err
	}

	for _, item := range metrics {
		if item.MType == consts.Gauge {
			t.UpdateGauge(item.ID, *item.Value)
		}
		if item.MType == consts.Counter {
			t.UpdateCounter(item.ID, *item.Delta)
		}
	}

	return nil
}

func (t MemStorage) Write() error {
	file, err := os.Create(t.storagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var metrics = make([]contracts.Metrics, 0)
	for name, value := range t.gaugeMetrics {
		metrics = append(metrics, contracts.Metrics{ID: name, MType: consts.Gauge, Value: &value})
	}
	for name, value := range t.counterMetrics {
		metrics = append(metrics, contracts.Metrics{ID: name, MType: consts.Counter, Delta: &value})
	}

	serialized, marshalErr := json.MarshalIndent(metrics, "", "   ")

	if marshalErr != nil {
		return marshalErr
	}

	_, writeErr := file.Write(serialized)

	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (t MemStorage) Ping() error {
	return nil
}

func (t MemStorage) UpdateMetrics(metrics []contracts.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, v := range metrics {
		if v.MType == consts.Gauge {
			err := t.UpdateGauge(v.ID, *v.Value)
			if err != nil {
				return err
			}
		}
		if v.MType == consts.Counter {
			err := t.UpdateCounter(v.ID, *v.Delta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
