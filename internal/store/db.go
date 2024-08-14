package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
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
	return conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn() // conn is a *pgx.Conn
		defer conn.Close(ctx)

		batch := &pgx.Batch{}
		for id, val := range m.Gauge {
			queryes := "INSERT INTO metrics (id, gauge) VALUES (@id, @val) ON CONFLICT (id) DO UPDATE SET gauge = @val"
			log.Printf("query: %s |%v %v\n", queryes, id, val)
			batch.Queue(queryes, pgx.NamedArgs{"id": id, "val": val})
		}
		for id, val := range m.Counter {
			queryes := "INSERT INTO metrics (id, counter) VALUES (@id, @val) ON CONFLICT (id) DO UPDATE SET counter = metrics.counter + @val"
			log.Printf("query: %s |%v %v\n", queryes, id, val)
			batch.Queue(queryes, pgx.NamedArgs{"id": id, "val": val})
		}
		br := conn.SendBatch(ctx, batch)

		var errs []error

		for i := 0; i < batch.Len(); i++ {
			ct, err := br.Exec()
			if err != nil {
				errs = append(errs, err)
			}
			if ct.RowsAffected() != 1 {
				errs = append(errs, fmt.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 1))
			}
		}

		err = br.Close()
		errs = append(errs, err)
		return errors.Join(errs...)
	})
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

	metrics := NewMemStorage()
	metrics.AddMetrics(m)

	log.Printf("AddMetrics %+v\n", metrics)

	retry := 0
	err := s.SetBatch(metrics)
	for needRetry(err) && retry < 3 {
		after := (retry+1)*2 - 1
		log.Printf("%s Retry %d ...", err.Error(), retry+1)
		time.Sleep(time.Duration(after) * time.Second)
		err = s.SetBatch(metrics)
		retry++
	}
	return err
}

func needRetry(err error) bool {
	var e *pgconn.PgError
	return errors.As(err, &e) && pgerrcode.IsConnectionException(e.Code)
}
func (s *DBStorage) GetMetrics(m *mtr.MetricsList) error {

	metricsID := make(map[string]bool)

	for _, v := range *m {
		metricsID[v.ID] = true
	}

	query := "SELECT id, gauge, counter FROM metrics where id in ("
	for k := range metricsID {
		query += fmt.Sprintf("'%s',", k)
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
	*m = mtr.MetricsList{}
	for rows.Next() {
		var nm mtr.Metrics
		err := rows.Scan(&nm.ID, &nm.Value, &nm.Delta)
		if err != nil {
			log.Println(err)
			return err
		}
		if nm.Delta == nil {
			nm.MType = GaugeName
		} else {
			nm.MType = CounterName
		}
		*m = append(*m, nm)
	}
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
