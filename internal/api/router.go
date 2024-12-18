package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xoxloviwan/go-monitor/internal/store"
	"golang.org/x/sync/errgroup"

	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	config "github.com/xoxloviwan/go-monitor/internal/config_server"
	grpcServ "github.com/xoxloviwan/go-monitor/internal/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Backuper is an interface for backing up data.
//
// It provides methods for backing up data and restoring data from a file.
type FileBackuper interface {
	SaveToFile(path string) error
	RestoreFromFile(path string) error
}

// Storage is alias for ReaderWriter.
type Storage interface {
	ReaderWriter
}

//go:generate mockgen -destination ./mock_router.go -package api github.com/xoxloviwan/go-monitor/internal/api Router

// Router interface for API server.
type Router interface {
	SetupRouter(ping gin.HandlerFunc, dbstore ReaderWriter, logLevel slog.Level, key []byte, privateKey *asc.PrivateKey, subnet *net.IPNet)
	Run(addr string) error
	Shutdown() error
}

// RunServer runs the API server with the given configuration.
//
// It sets up the routes, middleware, and logging, and starts the server.
func RunServer(r Router, cfg config.Config) error {
	var (
		s           Storage
		pingHandler gin.HandlerFunc
	)

	// Если DSN не пустой, то используем базу данных.
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
			return fmt.Errorf("create table error: %w", err)
		}
		s = dbs
	} else {
		// Если DSN пустой, то используем память.
		pingHandler = func(c *gin.Context) {
			c.Status(http.StatusOK)
		}
		s = store.NewMemStorage()
	}

	if cfg.Restore && cfg.FileStoragePath != "" {
		if b, ok := s.(FileBackuper); ok {
			if err := b.RestoreFromFile(cfg.FileStoragePath); err != nil {
				Log.Error("restore data error", "path", cfg.FileStoragePath, "error", err) // fix autotests for iter9 if file not exist
			}
		}
	}

	var err error
	var pKey *asc.PrivateKey
	if cfg.CryptoKey != "" {
		if pKey, err = asc.GetPrivateKey(cfg.CryptoKey); err != nil {
			return fmt.Errorf("get private key error: %w", err)
		}
	}

	var subnet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, subnet, err = net.ParseCIDR(cfg.TrustedSubnet)
	}
	if err != nil {
		return fmt.Errorf("parse subnet error: %w", err)
	}

	// Настраиваем маршруты.
	r.SetupRouter(pingHandler, s, slog.LevelInfo, []byte(cfg.Key), pKey, subnet)

	grpcL, err := net.Listen("tcp", ":2323")
	if err != nil {
		return fmt.Errorf("grpc listener error: %w", err)
	}
	Log.Info("Start listening gRPC on", "addr", grpcL.Addr())
	grpcS := grpcServ.NewGrpcServer(Log, []byte(cfg.Key), subnet)

	// Создаем канал для сигналов завершения.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var eg errgroup.Group
	eg.Go(func() error {
		// Ждем сигнала завершения.
		<-quit
		Log.Info("Shutdown Server...")
		signal.Stop(quit)
		close(quit) // Остановим периодическое сохранение данных в файл.
		// Завершаем работу сервера.
		grpcS.GracefulStop()
		return r.Shutdown()
	})

	// Запускаем периодическое сохранение данных в файл.
	var (
		b  FileBackuper
		ok bool
	)
	// Если объект реализует интерфейс FileBackuper, то сохраняем данные в файл.
	if b, ok = s.(FileBackuper); ok && cfg.StoreInterval > 0 && cfg.FileStoragePath != "" {
		eg.Go(func() error {
			backupTicker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer backupTicker.Stop()
			for {
				select {
				case <-backupTicker.C:
					if err := b.SaveToFile(cfg.FileStoragePath); err != nil {
						return fmt.Errorf("backup ticker data error: %w", err)
					}
				case <-quit:
					Log.Info("Shutdown backup ticker...")
					return nil
				}
			}
		})
	}

	grpcServ.SetupServer(grpcS, s)

	eg.Go(func() error {
		return grpcS.Serve(grpcL)
	})

	// Запускаем сервер http
	if err := r.Run(cfg.Address); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("run server error: %w", err)
		}
	}

	// Делем одноразовое сохранение данных в файл при завершении работы.
	if b, ok := s.(FileBackuper); ok && cfg.FileStoragePath != "" {
		if err := b.SaveToFile(cfg.FileStoragePath); err != nil {
			return fmt.Errorf("backup data error: %w", err)
		}
	}

	// Ждем завершения всех горутин и возвращаем ошибку, если она произошла.
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("goroutines error: %w", err)
	}
	Log.Info("Service stopped")
	return nil
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
func (r *RouterImpl) SetupRouter(ping gin.HandlerFunc, dbstore ReaderWriter, logLevel slog.Level, key []byte, privateKey *asc.PrivateKey, subnet *net.IPNet) {
	handler := newHandler(dbstore)
	r.Use(compressGzip())
	r.Use(logger(logLevel))
	if subnet != nil {
		r.Use(checkIP(subnet))
	}
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

// Run starts the server listening on the specified address.
func (r *RouterImpl) Run(addr string) error {
	r.srv = &http.Server{
		Addr:    addr,
		Handler: r.Handler(),
	}
	return r.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (r *RouterImpl) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.srv.Shutdown(ctx)
}
