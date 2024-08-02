package api

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunServer(address string, storePath string, restore bool, storeInterval int) error {
	r, s := SetupRouter()
	if restore {
		err := s.RestoreFromFile(storePath)
		if err != nil {
			log.Println(err)
		}
	}

	wasError := make(chan error)
	go func() {
		err := r.Run(address)
		if err != nil {
			wasError <- err
		}
	}()
	backupTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer backupTicker.Stop()
	for {
		select {
		case <-backupTicker.C:
			slog.Info(fmt.Sprintf("Backup to file %s ...", storePath))
			err := s.SaveToFile(storePath)
			if err != nil {
				return err
			}
		case err := <-wasError:
			return err
		}
	}
}

func SetupRouter() (*gin.Engine, *store.MemStorage) {
	store := store.NewMemStorage()
	ps := fmt.Sprintf("host=%s user=%s password=%s database=postgres sslmode=disable",
		`localhost:5432`, `postres`, `12345`)

	db, err := sql.Open("pgx", ps)
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	handler := NewHandler(store, db)
	r := gin.New()
	r.Use(logger())
	r.Use(compressGzip())
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)
	r.GET("/ping", handler.ping)
	return r, store
}
