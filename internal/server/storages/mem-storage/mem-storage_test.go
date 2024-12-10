package memstorage

import (
	"encoding/json"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
)

func TestUpdateCounter(t *testing.T) {
	storage := New("./metrics.json", true)
	storage.UpdateCounter("someCounterMetric", 500)
	if value := storage.GetCounters()["someCounterMetric"]; value != 500 {
		t.Errorf("UpdateCounter %d, want %d", value, 500)
	}
}

func TestUpdateGauge(t *testing.T) {
	storage := New("./metrics.json", true)
	storage.UpdateGauge("someGaugeMetric", 25.25)
	if value := storage.GetGauges()["someGaugeMetric"]; value != 25.25 {
		t.Errorf("UpdateGauge %g, want %g", value, 25.25)
	}
}

func BenchmarkUpdateGauge(b *testing.B) {
	storage := New("./metrics.json", true)
	for i := 0; i < b.N; i++ {
		storage.UpdateGauge("someGaugeMetric", 500)
	}
}

func BenchmarkUpdateGaugeSpeed(b *testing.B) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	storage := New("./metrics.json", true)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		storage.UpdateGauge("someGaugeMetric", 500)
	}
}

func createTestFile(content []contracts.Metrics) string {
	file, _ := os.CreateTemp("", "test_storage*.json")
	defer file.Close()

	data, _ := json.Marshal(content)
	file.Write(data)

	return file.Name()
}

func TestNew(t *testing.T) {
	storage := New("", false)

	if storage.gaugeMetrics == nil || storage.counterMetrics == nil {
		t.Fatalf("Metrics maps not initialized")
	}

	if storage.storagePath != "" {
		t.Errorf("Expected storagePath to be empty, got '%s'", storage.storagePath)
	}
}

func TestUpdateMetrics(t *testing.T) {
	storage := New("", false)

	metrics := []contracts.Metrics{
		{ID: "counter1", MType: consts.Counter, Delta: int64Pointer(100)},
		{ID: "gauge1", MType: consts.Gauge, Value: float64Pointer(99.99)},
	}

	err := storage.UpdateMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to update metrics: %v", err)
	}

	if len(storage.GetCounters()) != 1 || len(storage.GetGauges()) != 1 {
		t.Errorf("Expected 1 counter and 1 gauge, got %d counters and %d gauges",
			len(storage.GetCounters()), len(storage.GetGauges()))
	}
}

func TestRestore(t *testing.T) {
	metrics := []contracts.Metrics{
		{ID: "counter1", MType: consts.Counter, Delta: int64Pointer(50)},
		{ID: "gauge1", MType: consts.Gauge, Value: float64Pointer(11.11)},
	}
	filePath := createTestFile(metrics)
	defer os.Remove(filePath)

	storage := New(filePath, true)

	if len(storage.GetCounters()) != 1 || len(storage.GetGauges()) != 1 {
		t.Errorf("Expected metrics restored from file, but got %d counters and %d gauges",
			len(storage.GetCounters()), len(storage.GetGauges()))
	}
}

func TestWrite(t *testing.T) {
	filePath := createTestFile([]contracts.Metrics{})
	defer os.Remove(filePath)

	storage := New(filePath, false)
	_ = storage.UpdateGauge("gauge1", 20.20)
	_ = storage.UpdateCounter("counter1", 15)

	err := storage.Write()
	if err != nil {
		t.Fatalf("Failed to write metrics: %v", err)
	}

	content, _ := os.ReadFile(filePath)
	var metrics []contracts.Metrics
	_ = json.Unmarshal(content, &metrics)

	expected := []contracts.Metrics{
		{ID: "gauge1", MType: consts.Gauge, Value: float64Pointer(20.20)},
		{ID: "counter1", MType: consts.Counter, Delta: int64Pointer(15)},
	}

	if !reflect.DeepEqual(metrics, expected) {
		t.Errorf("Expected metrics %v, but got %v", expected, metrics)
	}
}

func TestGetCountValueByName(t *testing.T) {
	storage := New("", false)
	_, err := storage.GetCountValueByName("unknown")
	if err == nil {
		t.Errorf("Expected error for unknown counter, but got none")
	}
}

func TestGetGaugeValueByName(t *testing.T) {
	storage := New("", false)
	_, err := storage.GetGaugeValueByName("unknown")
	if err == nil {
		t.Errorf("Expected error for unknown gauge, but got none")
	}
}

func int64Pointer(v int64) *int64 {
	return &v
}

func float64Pointer(v float64) *float64 {
	return &v
}
