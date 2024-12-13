package base

import (
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
)

type Client struct {
	Addr      string
	Key       string
	LocalIP   string
	PublicKey *asc.PublicKey
}
