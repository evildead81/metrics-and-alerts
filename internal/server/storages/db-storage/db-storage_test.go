package dbstorage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	_ "github.com/lib/pq"
)

const (
	testDBConnString = "host=localhost user=zhabbarovrm password=zhabbarovrm dbname=zhabbarovrm sslmode=disable"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open("postgres", testDBConnString)
	if err != nil {
		log.Fatalf("Failed to connect to the test database: %v", err)
	}
	defer testDB.Close()

	if err := createTestTables(); err != nil {
		log.Fatalf("Failed to create test tables: %v", err)
	}

	code := m.Run()

	if err := dropTestTables(); err != nil {
		log.Printf("Failed to drop test tables: %v", err)
	}

	os.Exit(code)
}

func createTestTables() error {
	queries := []string{
		"CREATE TABLE IF NOT EXISTS gauges (id VARCHAR (50) PRIMARY KEY, value DOUBLE PRECISION);",
		"CREATE TABLE IF NOT EXISTS counters (id VARCHAR (50) PRIMARY KEY, value BIGINT);",
	}
	for _, query := range queries {
		if _, err := testDB.Exec(query); err != nil {
			return fmt.Errorf("failed to create test table: %w", err)
		}
	}
	return nil
}

func dropTestTables() error {
	queries := []string{
		"DROP TABLE IF EXISTS gauges;",
		"DROP TABLE IF EXISTS counters;",
	}
	for _, query := range queries {
		if _, err := testDB.Exec(query); err != nil {
			return fmt.Errorf("failed to drop test table: %w", err)
		}
	}
	return nil
}

func clearTables(db *sql.DB) error {
	queries := []string{
		"TRUNCATE TABLE gauges;",
		"TRUNCATE TABLE counters;",
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to clear table: %w", err)
		}
	}
	return nil
}
func TestIsCounterExists(t *testing.T) {
	storage := New(testDB)

	err := storage.UpdateCounter("counter_exists", 1)
	if err != nil {
		t.Fatalf("Failed to update counter: %v", err)
	}

	exists := storage.isCounterExists("counter_exists")
	if !exists {
		t.Error("Expected counter 'counter_exists' to exist, but it does not")
	}

	exists = storage.isCounterExists("counter_not_exists")
	if exists {
		t.Error("Expected counter 'counter_not_exists' to not exist, but it does")
	}
}

func TestIsGaugeExists(t *testing.T) {
	storage := New(testDB)

	err := storage.UpdateGauge("gauge_exists", 1.23)
	if err != nil {
		t.Fatalf("Failed to update gauge: %v", err)
	}

	exists := storage.isGaugeExists("gauge_exists")
	if !exists {
		t.Error("Expected gauge 'gauge_exists' to exist, but it does not")
	}

	exists = storage.isGaugeExists("gauge_not_exists")
	if exists {
		t.Error("Expected gauge 'gauge_not_exists' to not exist, but it does")
	}
}

func TestGetAllCounters(t *testing.T) {
	storage := New(testDB)

	if err := clearTables(testDB); err != nil {
		t.Fatalf("Failed to clear tables: %v", err)
	}

	_ = storage.UpdateCounter("counter1", 100)
	_ = storage.UpdateCounter("counter2", 200)

	counters := storage.GetCounters()
	if len(counters) != 2 || counters["counter1"] != 100 || counters["counter2"] != 200 {
		t.Errorf("Unexpected counters: %+v", counters)
	}
}

func TestGetAllGauges(t *testing.T) {
	storage := New(testDB)

	if err := clearTables(testDB); err != nil {
		t.Fatalf("Failed to clear tables: %v", err)
	}

	_ = storage.UpdateGauge("gauge1", 1.23)
	_ = storage.UpdateGauge("gauge2", 4.56)

	gauges := storage.GetGauges()
	if len(gauges) != 2 || gauges["gauge1"] != 1.23 || gauges["gauge2"] != 4.56 {
		t.Errorf("Unexpected gauges: %+v", gauges)
	}
}

func TestPing(t *testing.T) {
	storage := New(testDB)

	err := storage.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestUpdateCounter(t *testing.T) {
	storage := New(testDB)

	err := storage.UpdateCounter("test_counter", 10)
	if err != nil {
		t.Fatalf("Failed to update counter: %v", err)
	}

	value, err := storage.GetCountValueByName("test_counter")
	if err != nil {
		t.Fatalf("Failed to get counter value: %v", err)
	}
	if value != 10 {
		t.Errorf("Expected counter value 10, but got %d", value)
	}
}

func TestUpdateGauge(t *testing.T) {
	storage := New(testDB)

	err := storage.UpdateGauge("test_gauge", 42.42)
	if err != nil {
		t.Fatalf("Failed to update gauge: %v", err)
	}

	value, err := storage.GetGaugeValueByName("test_gauge")
	if err != nil {
		t.Fatalf("Failed to get gauge value: %v", err)
	}
	if value != 42.42 {
		t.Errorf("Expected gauge value 42.42, but got %f", value)
	}
}

func TestUpdateMetrics(t *testing.T) {
	storage := New(testDB)

	if err := clearTables(testDB); err != nil {
		t.Fatalf("Failed to clear tables: %v", err)
	}

	metrics := []contracts.Metrics{
		{
			ID:    "counter1",
			MType: consts.Counter,
			Delta: int64Pointer(10),
		},
		{
			ID:    "gauge1",
			MType: consts.Gauge,
			Value: float64Pointer(2.5),
		},
	}

	err := storage.UpdateMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to update metrics: %v", err)
	}

	counterValue, _ := storage.GetCountValueByName("counter1")
	if counterValue != 10 {
		t.Errorf("Expected counter1 value 10, but got %d", counterValue)
	}

	gaugeValue, _ := storage.GetGaugeValueByName("gauge1")
	if gaugeValue != 2.5 {
		t.Errorf("Expected gauge1 value 2.5, but got %f", gaugeValue)
	}

	metrics[0].Delta = int64Pointer(100)
	err = storage.UpdateMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to update metrics: %v", err)
	}

	counterValue, _ = storage.GetCountValueByName("counter1")
	if counterValue != 110 {
		t.Errorf("Expected counter1 value 110, but got %d", counterValue)
	}
}

func int64Pointer(v int64) *int64 {
	return &v
}

func float64Pointer(v float64) *float64 {
	return &v
}
