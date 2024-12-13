package http

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
	asc "github.com/xoxloviwan/go-monitor/internal/asymcrypto"
	"github.com/xoxloviwan/go-monitor/internal/clients/base"
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

type Client base.Client

// Send
//
// workerID - идентификатор потока
// msgs - список метрик
func (s *Client) Send(workerID int, msgs api.MetricsList) (err error) {
	cl := &http.Client{}

	url := "http://" + s.Addr + "/updates/"

	var body []byte
	body, err = easyjson.Marshal(&msgs)
	if err != nil {
		return err
	}
	var sessionKey []byte
	if s.PublicKey != nil {
		var err error
		sessionKey, body, err = asc.Encrypt(s.PublicKey, body)
		if err != nil {
			return err
		}
	}
	var sign string
	if s.Key != "" {
		sign, err = base.GetHash(body, s.Key)
		if err != nil {
			return err
		}
	}
	var gzbody []byte
	gzbody, err = base.CompressGzip(body)
	if err != nil {
		return err
	}
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(gzbody))
	if err != nil {
		return err
	}
	// net.IPAddr

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	if s.LocalIP != "" {
		req.Header.Set("X-Real-IP", s.LocalIP)
	}
	if sessionKey != nil {
		req.Header.Set("X-Key", hex.EncodeToString(sessionKey))
	}

	if s.Key != "" {
		req.Header.Set("HashSHA256", sign)
	}

	var response *http.Response
	retry := 0
	response, err = cl.Do(req)
	if err == nil && response.StatusCode != http.StatusOK {
		slog.Warn("Unexpected status code", "worker", workerID, "status_code", response.StatusCode)
	}
	for err != nil && retry < 3 {
		if response != nil {
			response.Body.Close()
		}
		after := (retry+1)*2 - 1
		time.Sleep(time.Duration(after) * time.Second)
		slog.Warn("Retry attempt", "worker", workerID, "error", err, "retry", retry+1)
		response, err = cl.Do(req)
		retry++
	}
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			closeErr = fmt.Errorf("could not close response body: %w", closeErr)
			err = errors.Join(err, closeErr)
		}
	}()

	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		return err
	}
	return nil
}
