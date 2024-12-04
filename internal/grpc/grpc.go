package grpcservice

import (
	// импортируем пакет со сгенерированными protobuf-файлами
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	mcv "github.com/xoxloviwan/go-monitor/internal/metrics_convert"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
		md, _ := metadata.FromIncomingContext(ctx)
		log.Info("RPC REQ", "method", info.FullMethod, "metadata", md)
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

/*
func decryptInterceptor(privateKey *asc.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if privateKey == nil {
			return handler(ctx, req)
		}
		body, ok := req.([]byte)
		if !ok {
			return nil, errors.New("not bytes")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("no metadata")
		}
		key := md.Get("X-Key")[0]
		encryptedSessionKey, err := hex.DecodeString(key)
		if err != nil {
			return nil, err
		}
		decrypted, err := asc.Decrypt(privateKey, encryptedSessionKey, body)
		if err != nil {
			return nil, err
		}
		return handler(ctx, decrypted)
	}
}
*/

func verifyHashInterceptor(key []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(key) == 0 {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "no metadata")
		}
		gotSignHeader := md.Get("HashSHA256")
		if len(gotSignHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "no hash")
		}
		gotSignHex := gotSignHeader[0]
		gotSign, err := hex.DecodeString(gotSignHex)
		if err != nil || len(gotSign) != sha256.Size {
			return nil, status.Errorf(codes.InvalidArgument, "invalid hash")
		}

		metrs, ok := req.(*pb.Metrics)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "wrong data type")
		}
		body := []byte(metrs.String())
		h := hmac.New(sha256.New, key)
		h.Write(body)
		sign := h.Sum(nil)
		if !hmac.Equal(sign, gotSign) {
			return nil, status.Errorf(codes.InvalidArgument, "hash sum not match %s %s", sign, gotSign)
		}
		return handler(ctx, req)
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

func NewGrpcServer(log logger, key []byte, subnet string) *grpc.Server {

	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc.UnaryServerInterceptor(logInterceptor(log)),
			grpc.UnaryServerInterceptor(verifyHashInterceptor(key)),
		),
	)
}

func SetupServer(grpcS *grpc.Server, store Storage) {
	pb.RegisterMetricsServiceServer(grpcS, &MetricsHandler{store: store})
}
