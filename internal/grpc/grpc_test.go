package grpcservice_test

import (
	"context"
	"crypto/sha256"
	"log"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	mock "github.com/xoxloviwan/go-monitor/internal/api/mock"
	grpcclient "github.com/xoxloviwan/go-monitor/internal/clients/grpc"
	grpcservice "github.com/xoxloviwan/go-monitor/internal/grpc"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func setup(t *testing.T) (*mock.MockReaderWriter, []byte) {
	lis = bufconn.Listen(bufSize)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	h := sha256.New()
	h.Write([]byte("secret"))
	key := h.Sum(nil)
	_, netIp, _ := net.ParseCIDR("192.168.1.0/26")
	s := grpcservice.NewGrpcServer(logger, key, netIp)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewMockReaderWriter(ctrl)
	grpcservice.SetupServer(s, m)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return m, key
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestAddMetrics(t *testing.T) {
	m, key := setup(t)
	cl := grpcclient.Client{
		Addr:    "passthrough://bufnet",
		LocalIP: "192.168.1.12",
		Key:     string(key),
	}
	metricItem := api.Metrics{
		ID:    "test",
		MType: "gauge",
	}
	var val float64 = 1.34
	metricItem.Value = &val
	msg := api.MetricsList{metricItem}
	m.EXPECT().AddMetrics(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	err := cl.SendWithOpts(1, msg, grpc.WithContextDialer(bufDialer))
	if err != nil {
		t.Fatalf("AddMetrics failed: %v", err)
	}
}
