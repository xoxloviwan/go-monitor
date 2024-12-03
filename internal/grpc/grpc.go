package grpcservice

import (
	// импортируем пакет со сгенерированными protobuf-файлами
	"context"

	mcv "github.com/xoxloviwan/go-monitor/internal/metrics_convert"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"

	"google.golang.org/grpc"
)

type Storage interface {
	AddMetrics(ctx context.Context, metrics *api.MetricsList) error
}

// MetricsHandler поддерживает все необходимые методы сервера.
type MetricsHandler struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServiceServer
	store Storage
}

type logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

func logInterceptor(log logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Info("RPC REQ", "method", info.FullMethod)
		log.Debug("RPC REQ", "method", info.FullMethod, "req", req)
		m, err := handler(ctx, req)
		if err != nil {
			log.Error("RPC", "method", info.FullMethod, "error", err)
		} else {
			log.Info("RPC RES", "method", info.FullMethod, "res", m)
		}
		return m, err
	}
}

// AddMetrics
func (srv *MetricsHandler) AddMetrics(ctx context.Context, in *pb.Metrics) (*pb.Response, error) {
	metrics := mcv.ConvMetricsInverse(in)
	var response pb.Response

	if err := srv.store.AddMetrics(ctx, metrics); err != nil {
		return nil, err
	}
	response.Success = true

	return &response, nil
}

func NewGrpcServer(log logger) *grpc.Server {
	return grpc.NewServer(grpc.UnaryInterceptor(logInterceptor(log)))
}

func SetupServer(grpcS *grpc.Server, store Storage) {
	pb.RegisterMetricsServiceServer(grpcS, &MetricsHandler{store: store})
}
