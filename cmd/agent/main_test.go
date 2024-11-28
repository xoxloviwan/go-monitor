package main

import (
	"testing"
)

func Test_getIP(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name: "get ip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIP()
			t.Logf("ip: %s", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
