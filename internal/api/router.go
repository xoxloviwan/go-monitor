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
	config "github.com/xoxloviwan/go-monitor/internal/config_server"
	"github.com/xoxloviwan/go-monitor/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunServer(cfg config.Config) error {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	r, s := SetupRouter(db)
	if cfg.Restore {
		err := s.RestoreFromFile(cfg.FileStoragePath)
		if err != nil {
			log.Println(err)
		}
	}

	wasError := make(chan error)
	go func() {
		err := r.Run(cfg.Address)
		if err != nil {
			wasError <- err
		}
	}()
	backupTicker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer backupTicker.Stop()
	for {
		select {
		case <-backupTicker.C:
			slog.Info(fmt.Sprintf("Backup to file %s ...", cfg.FileStoragePath))
			err := s.SaveToFile(cfg.FileStoragePath)
			if err != nil {
				return err
			}
		case err := <-wasError:
			return err
		}
	}
}

func SetupRouter(db *sql.DB) (*gin.Engine, *store.MemStorage) {
	store := store.NewMemStorage()
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

	return r, store
}
