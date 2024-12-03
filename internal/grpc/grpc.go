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

func NewGrpcServer() *grpc.Server {
	return grpc.NewServer()
}

func SetupServer(grpcS *grpc.Server, store Storage) {
	pb.RegisterMetricsServiceServer(grpcS, &MetricsHandler{store: store})
}
