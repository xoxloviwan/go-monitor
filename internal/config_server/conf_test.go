package config_test

import (
	"testing"

	conf "github.com/xoxloviwan/go-monitor/internal/config_server"
)

func TestInitConfig(t *testing.T) {
	cfg := conf.InitConfig()
	if cfg.Address != "" {
		t.Log("default address applied")
	}
}
