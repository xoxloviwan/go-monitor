package base

import (
	"testing"
)

func Test_GetIP(t *testing.T) {
	got, err := GetIP()
	t.Logf("ip: %s", got)
	if err != nil {
		t.Errorf("getIP() error = %v, wantErr %v", err, nil)
	}
}
