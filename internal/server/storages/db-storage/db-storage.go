package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/evildead81/metrics-and-alerts/internal/contracts"
	"github.com/evildead81/metrics-and-alerts/internal/server/consts"
	"github.com/evildead81/metrics-and-alerts/internal/server/storages"
)

type DBStorage struct {
	db      *sql.DB
	storage storages.Storage
}

func New(db *sql.DB) *DBStorage {
	storage := &DBStorage{
		db: db,
	}

	storage.initDB()
	return storage
}

func (s DBStorage) initDB() error {
	query := "CREATE TABLE IF NOT EXISTS gauges(" +
		"id VARCHAR (50) PRIMARY KEY," +
		"value DOUBLE PRECISION" +
		");" +
		"CREATE TABLE IF NOT EXISTS counters(" +
		"id VARCHAR (50) PRIMARY KEY," +
		"value INTEGER" +
		");"

	_, err := s.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) UpdateCounter(name string, value int64) error {
	query := `
	INSERT INTO counters (id, value) 
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE 
	SET value = counters.value + $2;
`
	_, err := s.db.Exec(query, name, value)

	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) UpdateGauge(name string, value float64) error {
	query := `
		INSERT INTO gauges (id, value) 
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE 
		SET value = EXCLUDED.value;
	`

	_, err := s.db.Exec(query, name, value)
	return err
}

func (s DBStorage) isCounterExists(name string) bool {
	var exists bool
	s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM counters WHERE id = $1)", name).Scan(&exists)
	return exists
}

func (s DBStorage) isGaugeExists(name string) bool {
	var exists bool
	s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM gauges WHERE id = $1)", name).Scan(&exists)
	return exists
}

func (s DBStorage) GetCounters() map[string]int64 {
	counters := make(map[string]int64)
	rows, err := s.db.Query("SELECT * FROM counters")
	if err != nil || rows.Err() != nil {
		return counters
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var value int64
		err = rows.Scan(&name, &value)
		if err != nil {
			return counters
		}
		counters[name] = value
	}

	return counters
}

func (s DBStorage) GetGauges() map[string]float64 {
	gauges := make(map[string]float64)
	rows, err := s.db.Query("SELECT * FROM gauges")
	if err != nil || rows.Err() != nil {
		return gauges
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var value float64
		err = rows.Scan(&name, &value)
		if err != nil {
			return gauges
		}
		gauges[name] = value
	}

	return gauges
}

func (s DBStorage) GetGaugeValueByName(name string) (float64, error) {
	var value float64
	err := s.db.QueryRow("SELECT value FROM gauges WHERE id = $1", name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	return value, err
}

func (s DBStorage) GetCountValueByName(name string) (int64, error) {
	var value int64
	err := s.db.QueryRow("SELECT value FROM counters WHERE id = $1", name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	return value, err
}

func (s DBStorage) Restore() error {
	return nil
}

func (s DBStorage) Write() error {
	return nil
}

func (s DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

func (s DBStorage) UpdateMetrics(metrics []contracts.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, v := range metrics {
		if v.MType == consts.Gauge {
			err = s.UpdateGauge(v.ID, *v.Value)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		if v.MType == consts.Counter {
			err = s.UpdateCounter(v.ID, *v.Delta)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	tx.Commit()
	return nil
}
