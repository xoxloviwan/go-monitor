package clients

import (
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	"github.com/xoxloviwan/go-monitor/internal/clients/grpc"
	"github.com/xoxloviwan/go-monitor/internal/clients/http"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

type Sender interface {
	Send(worker int, msgs api.MetricsList) error
}

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
