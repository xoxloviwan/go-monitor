package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v4"
)

func TestCreateTable(t *testing.T) {
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
