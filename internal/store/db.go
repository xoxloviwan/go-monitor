package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	pgx "github.com/jackc/pgx/v5"
	stdlib "github.com/jackc/pgx/v5/stdlib"
	"github.com/mailru/easyjson"
	mtr "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (s *DBStorage) CreateTable() error {
	var err error
	_, err = s.db.ExecContext(context.Background(), fmt.Sprintf(`CREATE TABLE IF NOT EXISTS metrics (
			id TEXT PRIMARY KEY,
			%s INTEGER,
			%s DOUBLE PRECISION)`,
		CounterName,
		GaugeName),
	)
	return err
}

func (s *DBStorage) SetBatch(m *MemStorage) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var conn *sql.Conn
	conn, err = s.db.Conn(ctx)
	if err != nil {
		return err
	}
	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn() // conn is a *pgx.Conn
		defer conn.Close(ctx)

		batch := &pgx.Batch{}
		for id, val := range m.Gauge {
			queryes := fmt.Sprintf("UPDATE metrics SET gauge = @%s WHERE id = @id", id)
			log.Printf("query: %s |%v %v\n", queryes, id, val)
			batch.Queue(queryes, pgx.NamedArgs{"id": id, id: val})
		}
		for id, val := range m.Counter {
			queryes := fmt.Sprintf("UPDATE metrics SET counter = @%s WHERE id = @id", id)
			log.Printf("query: %s |%v %v\n", queryes, id, val)
			batch.Queue(queryes, pgx.NamedArgs{"id": id, id: val})
		}
		br := conn.SendBatch(ctx, batch)
		err = br.Close()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Add(metricType string, metricName string, metricValue string) (err error) {

	query := fmt.Sprintf(`INSERT INTO metrics (id, %s, %s) VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET %s = $2, %s = $3;
	`,
		CounterName,
		GaugeName,
		CounterName,
		GaugeName,
	)
	if metricType == CounterName {
		_, err = s.db.ExecContext(context.Background(), query,
			metricName,
			metricValue,
			sql.NullFloat64{},
		)
	} else {
		_, err = s.db.ExecContext(context.Background(), query,
			metricName,
			sql.NullInt64{},
			metricValue,
		)
	}
	return err
}

func (s *DBStorage) AddMetrics(m *mtr.MetricsList) error {

	metrics := MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}

	for _, v := range *m {
		if v.MType == GaugeName {
			metrics.Gauge[v.ID] = *v.Value
		}
		if v.MType == CounterName {
			metrics.Counter[v.ID] = *v.Delta
		}
	}
	log.Printf("%+v\n", metrics)
	return s.SetBatch(&metrics)
}

func (s *DBStorage) GetMetrics(m *mtr.MetricsList) error {

	query := "SELECT id, gauge, counter FROM metrics where id in ("
	for _, v := range *m {
		query += fmt.Sprintf("'%s',", v.ID)
	}
	query = query[:len(query)-1]
	query += ")"
	log.Println(query)
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	log.Println("GetMetrics check 1")
	m = &mtr.MetricsList{}
	for rows.Next() {
		var nm mtr.Metrics
		err := rows.Scan(&nm.ID, &nm.Value, &nm.Delta)
		if err != nil {
			log.Println(err)
			return err
		}
		*m = append(*m, nm)
		log.Printf("GetMetrics check 2 %+v %v %v\n", nm, nm.Value, nm.Delta)
	}
	log.Printf("GetMetrics check 3 %+v\n", m)
	if err = rows.Err(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *DBStorage) Get(metricType string, metricName string) (string, bool) {
	var colName = GaugeName
	if metricType == CounterName {
		colName = CounterName
	}
	query := fmt.Sprintf(`SELECT %s FROM metrics WHERE id = $1`, colName)
	log.Println(query)
	row := s.db.QueryRowContext(context.Background(), query, metricName)
	var metricValue string
	err := row.Scan(&metricValue)
	if err != nil {
		log.Println(err)
		return "", false
	}
	return metricValue, true
}

func (s *DBStorage) String() string {
	query := "SELECT id, gauge, counter FROM metrics"
	log.Println(query)
	var err error
	var rows *sql.Rows
	rows, err = s.db.QueryContext(context.Background(), query)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer rows.Close()
	var ms mtr.MetricsList
	for rows.Next() {
		var m mtr.Metrics
		err := rows.Scan(&m.ID, &m.Value, &m.Delta)
		if err != nil {
			log.Println(err)
			return ""
		}
		if m.Delta == nil {
			m.MType = GaugeName
		} else {
			m.MType = CounterName
		}
		ms = append(ms, m)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return ""
	}
	str, err := easyjson.Marshal(ms)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(str)
}

func (s *DBStorage) RestoreFromFile(path string) error {
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
	err = s.SetBatch(&metrics)
	if err != nil {
		return err
	}
	return nil
}
