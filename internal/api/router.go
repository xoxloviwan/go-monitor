package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunServer(address string, storePath string, restore bool, storeInterval int) error {
	r, s, db := SetupRouter()
	defer db.Close()
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

func initDB() (*sql.DB, error) {
	ps := fmt.Sprintf("host=%s user=%s password=%s database=postgres sslmode=disable",
		`localhost`, `postgres`, `12345`)
	return sql.Open("pgx", ps)
}

func SetupRouter() (*gin.Engine, *store.MemStorage, *sql.DB) {
	store := store.NewMemStorage()
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	handler := NewHandler(store)
	r := gin.New()
	r.Use(logger())
	r.Use(compressGzip())
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)

	r.GET("/ping", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			log.Println("ping error:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	return r, store, db
}
