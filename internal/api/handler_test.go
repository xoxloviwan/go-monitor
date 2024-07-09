package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_update(t *testing.T) {
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
		{
			name:   "service post 404",
			url:    "/update/other",
			method: http.MethodPost,
			want: want{
				code:        404,
				response:    "",
				contentType: "plain/text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			urlSpl := strings.Split(tt.url, "/")
			if len(urlSpl) > 2 {
				req.SetPathValue("metricType", urlSpl[2])
			}
			if len(urlSpl) > 3 {
				req.SetPathValue("metricName", urlSpl[3])
			}
			if len(urlSpl) > 4 {
				req.SetPathValue("metricValue", urlSpl[4])
			}

			update(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
