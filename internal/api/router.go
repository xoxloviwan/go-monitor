package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	config "github.com/xoxloviwan/go-monitor/internal/config_server"
	"github.com/xoxloviwan/go-monitor/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunServer(cfg config.Config) error {
	var s DBStorage
	var pingHandler gin.HandlerFunc

	var mems Storage = store.NewMemStorage()

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return err
		}
		defer db.Close()
		pingHandler = func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				c.Error(err)
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
		}
		dbs := store.NewDBStorage(db)
		err = dbs.CreateTable()
		if err != nil {
			return fmt.Errorf("create table error: %v", err)
		}
		s = dbs
	} else {
		pingHandler = func(c *gin.Context) {
			c.Status(http.StatusOK)
		}
		s = mems
	}
	if cfg.Restore && cfg.FileStoragePath != "" {
		err := s.RestoreFromFile(cfg.FileStoragePath)
		if err != nil {
			return err
		}
	}
	r := SetupRouter(pingHandler, s)

	wasError := make(chan error)
	go func() {
		err := r.Run(cfg.Address)
		if err != nil {
			wasError <- err
		}
	}()

	if cfg.DatabaseDSN != "" || cfg.FileStoragePath == "" {
		err := <-wasError
		return err
	}

	// NewTicker бросает panic в случае, если интервал меньше нуля.
	if cfg.StoreInterval == 0 {
		for {
			select {
			case err := <-wasError:
				return err
			default:
				err := BackupData(mems, cfg.FileStoragePath)
				if err != nil {
					return err
				}
			}
		}
	}

	backupTicker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer backupTicker.Stop()
	for {
		select {
		case <-backupTicker.C:
			err := BackupData(mems, cfg.FileStoragePath)
			if err != nil {
				return err
			}
		case err := <-wasError:
			return err
		}
	}
}

type Backuper interface {
	SaveToFile(path string) error
}

type Restorer interface {
	RestoreFromFile(path string) error
}

type DBStorage interface {
	Restorer
	ReaderWriter
}

type Storage interface {
	Backuper
	DBStorage
}

func SetupRouter(ping gin.HandlerFunc, dbstore ReaderWriter) *gin.Engine {
	handler := NewHandler(dbstore)
	r := gin.New()
	r.Use(logger())
	r.Use(compressGzip())
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)

	r.GET("/ping", ping)

	return r
}

func BackupData(b Backuper, path string) error {
	slog.Info(fmt.Sprintf("Backup to file %s ...", path))
	return b.SaveToFile(path)
}
