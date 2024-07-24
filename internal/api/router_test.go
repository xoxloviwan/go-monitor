package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
				code:        http.StatusOK,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_200_counter",
			url:    "/update/counter/someMetric/23",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_404",
			url:    "/update/counter/someMetric/23",
			method: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_400",
			url:    "/update/other/some/23",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_404",
			url:    "/update/other",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
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
			defer res.Body.Close()

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
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
				code:        http.StatusOK,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_404_gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_200_counter",
			url:    "/value/counter/someMetric",
			method: http.MethodGet,
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_404",
			url:    "/value/other/some",
			method: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
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

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
		})
	}
}
