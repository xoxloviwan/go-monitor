package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type DbStorage struct {
	db *sql.DB
}

func NewDbStorage(db *sql.DB) *DbStorage {
	return &DbStorage{
		db: db,
	}
}

func (s *DbStorage) CreateTable() error {
	const createTableQuery = `CREATE TABLE IF NOT EXISTS "metrics" (
		"id" INTEGER PRIMARY KEY,
		"Alloc" DOUBLE PRECISION,
		"BuckHashSys" DOUBLE PRECISION,
		"Frees" DOUBLE PRECISION,
		"GCCPUFraction" DOUBLE PRECISION,
		"GCSys" DOUBLE PRECISION,
		"HeapAlloc" DOUBLE PRECISION,
		"HeapIdle" DOUBLE PRECISION,
		"HeapInuse" DOUBLE PRECISION,
		"HeapObjects" DOUBLE PRECISION,
		"HeapReleased" DOUBLE PRECISION,
		"HeapSys" DOUBLE PRECISION,
		"LastGC" DOUBLE PRECISION,
		"Lookups" DOUBLE PRECISION,
		"MCacheInuse" DOUBLE PRECISION,
		"MCacheSys" DOUBLE PRECISION,
		"MSpanInuse" DOUBLE PRECISION,
		"MSpanSys" DOUBLE PRECISION,
		"Mallocs" DOUBLE PRECISION,
		"NextGC" DOUBLE PRECISION,
		"NumForcedGC" DOUBLE PRECISION,
		"NumGC" DOUBLE PRECISION,
		"OtherSys" DOUBLE PRECISION,
		"PauseTotalNs" DOUBLE PRECISION,
		"StackInuse" DOUBLE PRECISION,
		"StackSys" DOUBLE PRECISION,
		"Sys" DOUBLE PRECISION,
		"TotalAlloc" DOUBLE PRECISION,
		"RandomValue" DOUBLE PRECISION,
		"PollCount" INTEGER
	)`
	var err error
	_, err = s.db.ExecContext(context.Background(), createTableQuery)
	return err
}

func (s *DbStorage) InitLine() error {
	var err error
	_, err = s.db.ExecContext(context.Background(), `TRUNCATE metrics`)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(context.Background(), `INSERT INTO metrics (
		"id",
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"PollCount"
	)
	VALUES (0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)`)
	return err
}

func (s *DbStorage) Add(metricType string, metricName string, metricValue string) (err error) {
	query := fmt.Sprintf(`UPDATE metrics SET "%s" = $1 WHERE id = 0`, metricName)
	log.Println(query)
	_, err = s.db.ExecContext(context.Background(), query, metricValue)
	return err
}

func (s *DbStorage) Get(metricType string, metricName string) (string, bool) {
	query := fmt.Sprintf(`SELECT "%s" FROM metrics WHERE id = 0`, metricName)
	log.Println(query)
	row := s.db.QueryRowContext(context.Background(), query)
	var metricValue string
	err := row.Scan(&metricValue)
	if err != nil {
		log.Println(err)
		return "", false
	}
	return metricValue, true
}

func (s *DbStorage) String() string {
	query := "SELECT * FROM metrics WHERE id = 0"
	log.Println(query)
	var err error
	var rows *sql.Rows
	rows, err = s.db.QueryContext(context.Background(), query)
	if err != nil {
		log.Println(err)
		return ""
	}
	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		log.Println(err)
		return ""
	}
	values := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i := range values {
		dest[i] = &values[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			log.Println(err)
			return ""
		}
	}
	var str = ""
	for i, colName := range cols {
		str += fmt.Sprintf("%s: %s\n", colName, values[i])
	}
	return str
}

func (s *DbStorage) SaveToFile(path string) error {
	// TODO
	return nil
}

func (s *DbStorage) RestoreFromFile(path string) error {
	// TODO
	return nil
}
