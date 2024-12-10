package main

import (
	"testing"
)

func Test_getIP(t *testing.T) {
	got, err := getIP()
	t.Logf("ip: %s", got)
	if err != nil {
		t.Errorf("getIP() error = %v, wantErr %v", err, nil)
	}
}
