package dbstorage

import (
	"database/sql"
	"strconv"

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

func (s *DBStorage) UpdateCounter(name string, value int64) {
	var query string
	currVal, err := s.GetCountValueByName(name)

	if err != nil {
		currVal = 0
	}

	if s.isCounterExists(name) {
		query = "UPDATE counters SET value = $2 WHERE id = $1;"
	} else {
		query = "INSERT INTO counters (id, value) VALUES ($1, $2);"
	}
	s.db.Exec(query, name, strconv.Itoa(int(currVal+value)))
}

func (s *DBStorage) UpdateGauge(name string, value float64) {
	var query string
	if s.isGaugeExists(name) {
		query = "UPDATE gauges SET value = $2 WHERE id = $1;"
	} else {
		query = "INSERT INTO gauges (id, value) VALUES ($1, $2);"
	}
	s.db.Exec(query, name, strconv.FormatFloat(value, 'f', -1, 64))
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (s DBStorage) GetCountValueByName(name string) (int64, error) {
	var value int64
	err := s.db.QueryRow("SELECT value FROM counters WHERE id = $1", name).Scan(&value)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (t DBStorage) Restore() error {
	return nil
}

func (t DBStorage) Write() error {
	return nil
}
