package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
