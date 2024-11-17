package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	asc "github.com/xoxloviwan/go-monitor/internal/asym_crypto"
	config "github.com/xoxloviwan/go-monitor/internal/config_server"
	"github.com/xoxloviwan/go-monitor/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// RunServer runs the API server with the given configuration.
//
// It sets up the routes, middleware, and logging, and starts the server.
func RunServer(r Router, cfg config.Config) error {
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
			slog.Warn("Restore failed", slog.Any("error", err.Error()))
		}
	}

	var pKey *asc.PrivateKey
	if cfg.CryptoKey != "" {
		var err error
		pKey, err = asc.GetPrivateKey(cfg.CryptoKey)
		if err != nil {
			return err
		}
	}

	r.SetupRouter(pingHandler, s, slog.LevelDebug, []byte(cfg.Key), pKey)

	wasError := make(chan error)
	go func() {
		err := r.Run(cfg.Address)
		if err != nil {
			wasError <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if cfg.DatabaseDSN != "" || cfg.FileStoragePath == "" {
		for {
			select {
			case err := <-wasError:
				return err
			case <-quit:
				r.Shutdown()
				return nil
			}
		}
	}

	// NewTicker бросает panic в случае, если интервал меньше нуля.
	if cfg.StoreInterval == 0 {
		for {
			select {
			case <-quit:
				r.Shutdown()
				return nil
			case err := <-wasError:
				return err
			default:
				err := backupData(mems, cfg.FileStoragePath)
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
			err := backupData(mems, cfg.FileStoragePath)
			if err != nil {
				return err
			}
		case <-quit:
			r.Shutdown()
			return nil
		case err := <-wasError:
			return err
		}
	}
}

// Backuper is an interface for backing up data.
//
// It provides a method for saving the data to a file.
type Backuper interface {
	SaveToFile(path string) error
}

// DBStorage is an interface for database storage.
//
// It provides methods for restoring data from a file and implementing the ReaderWriter interface.
type DBStorage interface {
	RestoreFromFile(path string) error
	ReaderWriter
}

// Storage is an interface that combines Backuper and DBStorage.
//
// It provides methods for backing up data and restoring data from a file.
type Storage interface {
	Backuper
	DBStorage
}

//go:generate mockgen -destination ./mock_router.go -package api github.com/xoxloviwan/go-monitor/internal/api Router

// Router interface for API server.
type Router interface {
	SetupRouter(ping gin.HandlerFunc, dbstore ReaderWriter, logLevel slog.Level, key []byte, privateKey *asc.PrivateKey)
	Run(addr string) error
	Shutdown()
}

// RouterImpl is wrap for gin.Engine and http.Server
type RouterImpl struct {
	*gin.Engine
	srv *http.Server
}

// NewRouter returns a new Router instance.
func NewRouter() *RouterImpl {
	return &RouterImpl{gin.New(), nil}
}

// SetupRouter sets up routes and middleware.
//
// The engine is initialized with the given ping handler, store, log level, and key.
func (r *RouterImpl) SetupRouter(ping gin.HandlerFunc, dbstore ReaderWriter, logLevel slog.Level, key []byte, privateKey *asc.PrivateKey) {
	handler := newHandler(dbstore)
	r.Use(compressGzip())
	r.Use(logger(logLevel))
	if len(key) > 0 {
		r.Use(verifyHash(key))
	}
	if privateKey != nil {
		r.Use(decryptBody(privateKey))
	}
	r.POST("/update/:metricType/:metricName/:metricValue", handler.update)
	r.POST("/update/", handler.updateJSON)
	r.POST("/updates/", handler.updateJSON)
	r.GET("/value/:metricType/:metricName", handler.value)
	r.POST("/value/", handler.valueJSON)
	r.GET("/", handler.list)

	r.GET("/ping", ping)
}

func backupData(b Backuper, path string) error {
	slog.Info(fmt.Sprintf("Backup to file %s ...", path))
	return b.SaveToFile(path)
}

func (r *RouterImpl) Run(addr string) error {
	r.srv = &http.Server{
		Addr:    addr,
		Handler: r.Handler(),
	}
	return r.srv.ListenAndServe()
}

func (r *RouterImpl) Shutdown() {
	slog.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.srv.Shutdown(ctx); err != nil {
		slog.Error("Server Shutdown:", slog.Any("error", err))
		return
	}
	select {
	case <-ctx.Done():
		slog.Info("timeout of 5 seconds.")
	}
	slog.Info("Server exiting")
}
