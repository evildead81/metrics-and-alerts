package memstorage

import (
	"math/rand"
	"testing"
	"time"
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
