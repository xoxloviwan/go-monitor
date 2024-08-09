package store

import "database/sql"

type DbStorage struct {
	db *sql.DB
}

func NewDbStorage(db *sql.DB) *DbStorage {
	return &DbStorage{
		db: db,
	}
}

func (s *DbStorage) Add(metricType string, metricName string, metricValue string) (err error) {
	// TODO
	return nil
}

func (s *DbStorage) Get(metricType string, metricName string) (string, bool) {
	// TODO
	return "", false
}

func (s *DbStorage) String() string {
	// TODO
	return ""
}

func (s *DbStorage) SaveToFile(path string) error {
	// TODO
	return nil
}

func (s *DbStorage) RestoreFromFile(path string) error {
	// TODO
	return nil
}
