package store

import (
	"context"
	"fmt"
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	mtr "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

func ExampleDBStorage_GetMetrics() {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewDBStorage(db)

	metricsList := mtr.MetricsList{
		{ID: "metric1"},
		{ID: "metric2"},
	}

	rows := sqlmock.NewRows([]string{"id", "value", "delta"}).
		AddRow("metric1", 10.5, nil).
		AddRow("metric2", nil, int64(20))

	mock.ExpectQuery("SELECT id, gauge, counter FROM metrics where id in").
		WillReturnRows(rows)

	metrics, err := store.GetMetrics(context.Background(), metricsList)
	if err != nil {
		log.Fatal(err)
	}

	json, err := metrics.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))

	// Output:
	// [{"id":"metric1","type":"gauge","value":10.5},{"id":"metric2","type":"counter","delta":20}]
}
