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
	pingHandler := func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			c.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}
	//s := store.NewMemStorage()
	s := store.NewDbStorage(db)
	err = s.CreateTable()
	if err != nil {
		return fmt.Errorf("create table error: %v", err)
	}
	err = s.InitLine()
	if err != nil {
		return fmt.Errorf("create init line error: %v", err)
	}
	r := SetupRouter(pingHandler, s)
	if cfg.Restore {
		RestoreData(s, cfg.FileStoragePath)
	}

	wasError := make(chan error)
	go func() {
		err := r.Run(cfg.Address)
		if err != nil {
			wasError <- err
		}
	}()

	// NewTicker бросает panic в случае, если интервал меньше нуля.
	if cfg.StoreInterval == 0 {
		for {
			select {
			case err := <-wasError:
				return err
			default:
				err := BackupData(s, cfg.FileStoragePath)
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
			err := BackupData(s, cfg.FileStoragePath)
			if err != nil {
				return err
			}
		case err := <-wasError:
			return err
		}
	}
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

type Backuper interface {
	SaveToFile(path string) error
	RestoreFromFile(path string) error
}

func RestoreData(b Backuper, path string) {
	err := b.RestoreFromFile(path)
	if err != nil {
		log.Println(err)
	}
}

func BackupData(b Backuper, path string) error {
	slog.Info(fmt.Sprintf("Backup to file %s ...", path))
	return b.SaveToFile(path)
}
