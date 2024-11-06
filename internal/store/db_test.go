package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v4"
)

func TestCreateTable(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewDBStorage(db)
	res := sqlmock.NewErrorResult(nil)

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS`).WillReturnResult(res)

	err = store.CreateTable()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetBatchPgx(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close(ctx)

	memstore := &MemStorage{
		Gauge: map[string]float64{
			"item1": 1.1,
			"item2": 1.2,
			"item3": 1.3,
		},
		Counter: map[string]int64{
			"item1": 1,
			"item2": 2,
		},
	}

	eb := mock.ExpectBatch()
	eb.ExpectExec("INSERT INTO metrics").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	eb.ExpectExec("INSERT INTO metrics").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	eb.ExpectExec("INSERT INTO metrics").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	eb.ExpectExec("INSERT INTO metrics").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	eb.ExpectExec("INSERT INTO metrics").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	err = setBatchPgx(ctx, mock, memstore)
	if err != nil {
		t.Error(err)
	}
}

func TestString(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewDBStorage(db)
	rows := sqlmock.NewRows([]string{"id", "gauge", "counter"}).
		AddRow("item1", 2.5, nil).
		AddRow("item2", nil, 4)

	mock.ExpectQuery("SELECT id, gauge, counter FROM metrics").WillReturnRows(rows)

	str := store.String()
	fmt.Println(str)
	if str == "" {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewDBStorage(db)
	rows := sqlmock.NewRows([]string{"counter"}).AddRow("4")

	mock.ExpectQuery("SELECT counter FROM metrics").WillReturnRows(rows)

	str, ok := store.Get("counter", "item2")
	if !ok {
		t.Errorf("item not found")
	}
	if str != "4" {
		t.Errorf("wrong value of item: %s", str)
	}
}
