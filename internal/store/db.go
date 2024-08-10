package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mailru/easyjson"
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

func (s *DbStorage) SetLine(m *MemStorage) error {
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
	VALUES (0, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)`,
		m.Gauge["Alloc"],
		m.Gauge["BuckHashSys"],
		m.Gauge["Frees"],
		m.Gauge["GCCPUFraction"],
		m.Gauge["GCSys"],
		m.Gauge["HeapAlloc"],
		m.Gauge["HeapIdle"],
		m.Gauge["HeapInuse"],
		m.Gauge["HeapObjects"],
		m.Gauge["HeapReleased"],
		m.Gauge["HeapSys"],
		m.Gauge["LastGC"],
		m.Gauge["Lookups"],
		m.Gauge["MCacheInuse"],
		m.Gauge["MCacheSys"],
		m.Gauge["MSpanInuse"],
		m.Gauge["MSpanSys"],
		m.Gauge["Mallocs"],
		m.Gauge["NextGC"],
		m.Gauge["NumForcedGC"],
		m.Gauge["NumGC"],
		m.Gauge["OtherSys"],
		m.Gauge["PauseTotalNs"],
		m.Gauge["StackInuse"],
		m.Gauge["StackSys"],
		m.Gauge["Sys"],
		m.Gauge["TotalAlloc"],
		m.Gauge["RandomValue"],
		m.Counter["PollCount"])
	return err
}

func (s *DbStorage) GetLine(m *MemStorage) error {
	query := `SELECT
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
	FROM metrics WHERE id = 0`

	row := s.db.QueryRowContext(context.Background(), query)

	type Metrics struct {
		Alloc         float64
		BuckHashSys   float64
		Frees         float64
		GCCPUFraction float64
		GCSys         float64
		HeapAlloc     float64
		HeapIdle      float64
		HeapInuse     float64
		HeapObjects   float64
		HeapReleased  float64
		HeapSys       float64
		LastGC        float64
		Lookups       float64
		MCacheInuse   float64
		MCacheSys     float64
		MSpanInuse    float64
		MSpanSys      float64
		Mallocs       float64
		NextGC        float64
		NumForcedGC   float64
		NumGC         float64
		OtherSys      float64
		PauseTotalNs  float64
		StackInuse    float64
		StackSys      float64
		Sys           float64
		TotalAlloc    float64
		RandomValue   float64
		PollCount     int64
	}

	mm := &Metrics{}

	err := row.Scan(
		&mm.Alloc,
		&mm.BuckHashSys,
		&mm.Frees,
		&mm.GCCPUFraction,
		&mm.GCSys,
		&mm.HeapAlloc,
		&mm.HeapIdle,
		&mm.HeapInuse,
		&mm.HeapObjects,
		&mm.HeapReleased,
		&mm.HeapSys,
		&mm.LastGC,
		&mm.Lookups,
		&mm.MCacheInuse,
		&mm.MCacheSys,
		&mm.MSpanInuse,
		&mm.MSpanSys,
		&mm.Mallocs,
		&mm.NextGC,
		&mm.NumForcedGC,
		&mm.NumGC,
		&mm.OtherSys,
		&mm.PauseTotalNs,
		&mm.StackInuse,
		&mm.StackSys,
		&mm.Sys,
		&mm.TotalAlloc,
		&mm.RandomValue,
		&mm.PollCount)

	if err != nil {
		return err
	}

	m.Gauge["Alloc"] = mm.Alloc
	m.Gauge["BuckHashSys"] = mm.BuckHashSys
	m.Gauge["Frees"] = mm.Frees
	m.Gauge["GCCPUFraction"] = mm.GCCPUFraction
	m.Gauge["GCSys"] = mm.GCSys
	m.Gauge["HeapAlloc"] = mm.HeapAlloc
	m.Gauge["HeapIdle"] = mm.HeapIdle
	m.Gauge["HeapInuse"] = mm.HeapInuse
	m.Gauge["HeapObjects"] = mm.HeapObjects
	m.Gauge["HeapReleased"] = mm.HeapReleased
	m.Gauge["HeapSys"] = mm.HeapSys
	m.Gauge["LastGC"] = mm.LastGC
	m.Gauge["Lookups"] = mm.Lookups
	m.Gauge["MCacheInuse"] = mm.MCacheInuse
	m.Gauge["MCacheSys"] = mm.MCacheSys
	m.Gauge["MSpanInuse"] = mm.MSpanInuse
	m.Gauge["MSpanSys"] = mm.MSpanSys
	m.Gauge["Mallocs"] = mm.Mallocs
	m.Gauge["NextGC"] = mm.NextGC
	m.Gauge["NumForcedGC"] = mm.NumForcedGC
	m.Gauge["NumGC"] = mm.NumGC
	m.Gauge["OtherSys"] = mm.OtherSys
	m.Gauge["PauseTotalNs"] = mm.PauseTotalNs
	m.Gauge["StackInuse"] = mm.StackInuse
	m.Gauge["StackSys"] = mm.StackSys
	m.Gauge["Sys"] = mm.Sys
	m.Gauge["TotalAlloc"] = mm.TotalAlloc
	m.Gauge["RandomValue"] = mm.RandomValue
	m.Counter["PollCount"] = mm.PollCount
	return nil
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
	defer rows.Close()
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
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return ""
	}
	var str = ""
	for i, colName := range cols {
		str += fmt.Sprintf("%s: %s\n", colName, values[i])
	}
	return str
}

func (s *DbStorage) SaveToFile(path string) error {
	var metrics = MemStorage{
		Gauge:   make(Gauge),
		Counter: make(Counter),
	}
	var err error
	var data []byte
	err = s.GetLine(&metrics)
	if err != nil {
		return err
	}
	data, err = easyjson.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *DbStorage) RestoreFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var metrics MemStorage
	log.Println(string(data))
	err = easyjson.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}
	err = s.SetLine(&metrics)
	if err != nil {
		return err
	}
	return nil
}
