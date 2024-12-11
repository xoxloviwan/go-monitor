package grpcservice_test

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	mock "github.com/xoxloviwan/go-monitor/internal/api/mock"
	grpcservice "github.com/xoxloviwan/go-monitor/internal/grpc"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func setup(t *testing.T) *mock.MockReaderWriter {
	lis = bufconn.Listen(bufSize)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	s := grpcservice.NewGrpcServer(logger, nil, nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewMockReaderWriter(ctrl)
	grpcservice.SetupServer(s, m)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return m
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestAddMetrics(t *testing.T) {
	m := setup(t)
	ctx := context.Background()
	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewMetricsServiceClient(conn)

	metricItem := &pb.Metric{
		Id:    "test",
		Type:  "gauge",
		Value: 1.34,
		Delta: 0,
	}
	msg := &pb.Metrics{
		Metrics: []*pb.Metric{metricItem},
	}
	m.EXPECT().AddMetrics(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	resp, err := client.AddMetrics(ctx, msg)
	if err != nil {
		t.Fatalf("AddMetrics failed: %v", err)
	}
	log.Printf("Response: %+v", resp)
}
