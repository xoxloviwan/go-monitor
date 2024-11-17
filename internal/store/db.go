package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	stdlib "github.com/jackc/pgx/v5/stdlib"
	"github.com/mailru/easyjson"
	mtr "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

// DBStorage is a database storage implementation.
//
// It provides methods for creating a table, setting batch data, adding metrics, getting metrics, and restoring data from a file.
type DBStorage struct {
	db *sql.DB
}

// NewDBStorage returns a new DBStorage instance.
//
// The instance is initialized with the given sql.DB.
func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

// CreateTable creates a table in the database if it does not exist.
//
// The table is created with the given column names and types.
func (s *DBStorage) CreateTable() error {
	var err error
	_, err = s.db.ExecContext(context.Background(), fmt.Sprintf(`CREATE TABLE IF NOT EXISTS metrics (
			id TEXT PRIMARY KEY,
			%s BIGINT,
			%s DOUBLE PRECISION)`,
		CounterName,
		GaugeName),
	)
	return err
}

// setBatch sets batch data in the database.
//
// The data is set in the given context with the given timeout.
func setBatch(parent context.Context, db *sql.DB, m *MemStorage) error {

	ctx, cancel := context.WithTimeout(parent, 120*time.Second)
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	return conn.Raw(func(driverConn any) error {
		conn := driverConn.(*stdlib.Conn).Conn() // conn is a *pgx.Conn
		defer conn.Close(ctx)
		return setBatchPgx(ctx, conn, m)
	})
}

// PgxIface represents an interface for sending batches of SQL statements to a PostgreSQL database using the pgx library.
// There is no way we can hijack the real pgx.Conn for pgxmock.
//
// Links:
// * https://github.com/pashagolub/pgxmock/issues/20
// * https://github.com/jackc/pgx/pull/1996#issuecomment-2090089072
type PgxIface interface {
	SendBatch(ctx context.Context, b *pgx.Batch) (br pgx.BatchResults)
}

func setBatchPgx(ctx context.Context, conn PgxIface, m *MemStorage) (err error) {
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

	defer func() error {
		err = br.Close()
		if err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}()

	for i := 0; i < batch.Len(); i++ {
		ct, err := br.Exec()
		if err != nil {
			errs = append(errs, err)
		}
		if ct.RowsAffected() != 1 {
			errs = append(errs, fmt.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 1))
		}
	}

	return errors.Join(errs...)
}

// Add adds a metric to the database.
//
// The metric is added with the given type, name, and value.
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

// AddMetrics adds multiple metrics to the database.
//
// The metrics are added with the given context and metrics list.
func (s *DBStorage) AddMetrics(ctx context.Context, m *mtr.MetricsList) error {

	metrics := NewMemStorage()
	metrics.AddMetrics(ctx, m)

	retry := 0
	err := setBatch(ctx, s.db, metrics)
	for needRetry(err) && retry < 3 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			after := (retry+1)*2 - 1
			slog.Error(fmt.Sprintf("%s Retry %d ...", err.Error(), retry+1))
			time.Sleep(time.Duration(after) * time.Second)
			err = setBatch(ctx, s.db, metrics)
			retry++
		}
	}
	return err
}

func needRetry(err error) bool {
	var e *pgconn.PgError
	return errors.As(err, &e) && pgerrcode.IsConnectionException(e.Code)
}

func makeQueryString(metricsList mtr.MetricsList) (string, error) {
	var err error
	b := bytes.NewBufferString("SELECT id, gauge, counter FROM metrics where id in (")
	for k, v := range metricsList {
		if (k) == len(metricsList)-1 {
			_, err = fmt.Fprintf(b, "'%s')", v.ID)
		} else {
			_, err = fmt.Fprintf(b, "'%s',", v.ID)
		}
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// GetMetrics gets metrics from the database.
//
// The metrics are retrieved with the given context and metrics list.
func (s *DBStorage) GetMetrics(ctx context.Context, metricsList mtr.MetricsList) (mtr.MetricsList, error) {
	query, err := makeQueryString(metricsList)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(query)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	metricsListWithValues := mtr.MetricsList{}
	for rows.Next() {
		var nm mtr.Metrics
		err := rows.Scan(&nm.ID, &nm.Value, &nm.Delta)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if nm.Delta == nil {
			nm.MType = GaugeName
		} else {
			nm.MType = CounterName
		}
		metricsListWithValues = append(metricsListWithValues, nm)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return metricsListWithValues, nil
}

// Get gets a metric from the database.
//
// The metric is retrieved with the given type and name.
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

// String returns a string representation of the DBStorage instance.
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

// RestoreFromFile restores data from a file.
//
// The data is restored from the given file path.
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
	err = setBatch(context.Background(), s.db, &metrics)
	if err != nil {
		return err
	}
	return nil
}
