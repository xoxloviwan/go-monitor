package clients

import (
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	"github.com/xoxloviwan/go-monitor/internal/clients/grpc"
	"github.com/xoxloviwan/go-monitor/internal/clients/http"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

// Sender is an interface that defines the contract for sending metrics data.
// The Send method is used to send a list of metrics for a given worker.
// The method returns an error if the send operation fails.
type Sender interface {
	Send(worker int, msgs api.MetricsList) error
}

// NewSender creates a new Sender instance based on the provided configuration.
// If grpcFlag is true, it returns a gRPC-based Sender implementation.
// Otherwise, it returns an HTTP-based Sender implementation.
// The Sender implementation is responsible for sending metrics data to the monitoring system.
func NewSender(grpcFlag bool, addr string, key string, localIP string, publicKey *asc.PublicKey) Sender {
	if grpcFlag {
		return &grpc.Client{
			Addr:      addr,
			Key:       key,
			LocalIP:   localIP,
			PublicKey: publicKey,
		} // не знаю как избежать здесь дублирования при инициализации разных структур с одинаковым набором полей
	}
	return &http.Client{
		Addr:      addr,
		Key:       key,
		LocalIP:   localIP,
		PublicKey: publicKey,
	}
}
