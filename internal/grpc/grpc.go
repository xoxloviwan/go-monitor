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

// MetricsServer поддерживает все необходимые методы сервера.
type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServiceServer
	store Storage
}

// AddUser реализует интерфейс добавления пользователя.
func (srv *MetricsServer) AddMetrics(ctx context.Context, in *pb.Metrics) (*pb.Response, error) {
	metrics := mcv.ConvMetricsInverse(in)
	var response pb.Response

	if err := srv.store.AddMetrics(ctx, metrics); err != nil {
		return nil, err
	}

	return &response, nil
}

func registerGrpcService(grpcS *grpc.Server, store Storage) {
	pb.RegisterMetricsServiceServer(grpcS, &MetricsServer{store: store})
}
