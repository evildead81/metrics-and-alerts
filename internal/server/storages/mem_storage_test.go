package storages

import (
	"testing"
)

func TestUpdateCounter(t *testing.T) {
	storage := New()
	storage.UpdateCounter("someCounterMetric", 500)
	if value := storage.GetCounters()["someCounterMetric"]; value != 500 {
		t.Errorf("UpdateCounter %d, want %d", value, 500)
	}
}

func TestUpdateGauge(t *testing.T) {
	storage := New()
	storage.UpdateGauge("someCounterMetric", 25.25)
	if value := storage.GetGauges()["someCounterMetric"]; value != 25.25 {
		t.Errorf("UpdateCounter %g, want %g", value, 25.25)
	}
}
