package base

import (
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
)

// Client represents a client connection to the server.
type Client struct {
	Addr      string
	Key       string
	LocalIP   string
	PublicKey *asc.PublicKey
}
