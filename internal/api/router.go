package api

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"
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
	handler := NewHandler(store)
	r := gin.New()
	r.Use(logger())
	r.Use(compressGzip())
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)
	return r, store
}
