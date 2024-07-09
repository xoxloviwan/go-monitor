package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xoxloviwan/go-monitor/internal/store"
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

var hdl = &Handler{
	store: store.NewMemStorage(),
}

func Test_update(t *testing.T) {

	tests := testcase{
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

			hdl.update(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func Test_value(t *testing.T) {
	tests := testcase{
		{
			name:   "service get 200 gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service post 400 gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service get 200 counter",
			url:    "/value/counter/someMetric",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service get 404",
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

			hdl.value(w, req)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
