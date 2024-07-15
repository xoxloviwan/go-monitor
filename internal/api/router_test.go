package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type want struct {
	code        int
	response    string
	contentType string
}

type testcase []struct {
	name   string
	url    string
	method string
	want   want
}

var router = SetupRouter()

func Test_update(t *testing.T) {

	tests := testcase{
		{
			name:   "service_post_200_gauge",
			url:    "/update/gauge/someMetric/23.4",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_200_counter",
			url:    "/update/counter/someMetric/23",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_400",
			url:    "/update/counter/someMetric/23",
			method: http.MethodGet,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_400",
			url:    "/update/other/some/23",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_404",
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

			router.ServeHTTP(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func Test_value(t *testing.T) {
	tests := testcase{
		{
			name:   "service_get_200_gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_400_gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_200_counter",
			url:    "/value/counter/someMetric",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_404",
			url:    "/value/other/some",
			method: http.MethodGet,
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

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
