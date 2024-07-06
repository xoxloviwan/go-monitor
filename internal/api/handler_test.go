package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "service post 200 gauge",
			url:    "/update/gauge/someMetric/23.4",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service post 200 counter",
			url:    "/update/counter/someMetric/23",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service get 400",
			url:    "/update/counter/someMetric/23",
			method: http.MethodGet,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service post 400",
			url:    "/update/other/some/23",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			req.SetPathValue("metricType", strings.Split(tt.url, "/")[2])
			req.SetPathValue("metricName", strings.Split(tt.url, "/")[3])
			req.SetPathValue("metricValue", strings.Split(tt.url, "/")[4])

			update(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
