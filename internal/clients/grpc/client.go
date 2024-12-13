package grpc

import (
	"context"
	"log/slog"

	"github.com/xoxloviwan/go-monitor/internal/clients/base"
	mcv "github.com/xoxloviwan/go-monitor/internal/metrics_convert"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

type Client base.Client

func (s *Client) Send(worker int, msgs api.MetricsList) (err error) {
	return s.SendWithOpts(worker, msgs)
}

func (s *Client) SendWithOpts(worker int, msgs api.MetricsList, opts ...grpc.DialOption) (err error) {

	slog.Info("gRPC worker got task", "worker", worker)
	// устанавливаем соединение с сервером
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(s.Addr, opts...)

	if err != nil {
		return err
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа MetricsServiceClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricsServiceClient(conn)
	md := metadata.New(map[string]string{
		"X-Real-IP": s.LocalIP,
	})
	metrs := mcv.ConvMetrics(msgs)
	if s.Key != "" {
		msg := metrs.String()
		sign, err := base.GetHash([]byte(msg), s.Key)
		if err != nil {
			return err
		}
		md.Set("HashSHA256", sign)
	}
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	MetricsResponse, err := c.AddMetrics(ctx, metrs, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return err
	}
	slog.Info("gRPC worker got response", "worker", worker, "response", MetricsResponse)
	return nil
}
