package clients

import (
	"reflect"
	"testing"

	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	"github.com/xoxloviwan/go-monitor/internal/clients/grpc"
	"github.com/xoxloviwan/go-monitor/internal/clients/http"
)

func TestNewSender(t *testing.T) {
	type args struct {
		grpcFlag  bool
		addr      string
		key       string
		localIP   string
		publicKey *asc.PublicKey
	}
	tests := []struct {
		name string
		args args
		want Sender
	}{
		{
			name: "make http client",
			args: args{
				grpcFlag:  false,
				addr:      "localhost:8080",
				key:       "",
				localIP:   "",
				publicKey: nil,
			},
			want: &http.Client{
				Addr:      "localhost:8080",
				Key:       "",
				LocalIP:   "",
				PublicKey: nil,
			},
		},
		{
			name: "make grpc client",
			args: args{
				grpcFlag:  true,
				addr:      "localhost:8080",
				key:       "",
				localIP:   "",
				publicKey: nil,
			},
			want: &grpc.Client{
				Addr:      "localhost:8080",
				Key:       "",
				LocalIP:   "",
				PublicKey: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSender(tt.args.grpcFlag, tt.args.addr, tt.args.key, tt.args.localIP, tt.args.publicKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSender() = %v, want %v", got, tt.want)
			}
		})
	}
}
