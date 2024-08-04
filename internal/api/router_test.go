package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type want struct {
	code        int
	contentType string
}

type testcase struct {
	name   string
	url    string
	method string
	want   want
}

type testcases []testcase

type testcasesWithBody []struct {
	testcase
	reqBody  string
	resBody  string
	wantBody string
}

func ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

var router, _ = SetupRouter(ping)

func Test_update(t *testing.T) {

	tests := testcases{
		{
			name:   "service_post_200_gauge",
			url:    "/update/gauge/someMetric/23.4",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_200_counter",
			url:    "/update/counter/someMetric/23",
			method: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_404",
			url:    "/update/counter/someMetric/23",
			method: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_400",
			url:    "/update/other/some/23",
			method: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_404",
			url:    "/update/other",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
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
	tests := testcases{
		{
			name:   "service_get_200_gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodGet,
			want: want{
				code:        http.StatusOK,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_post_404_gauge",
			url:    "/value/gauge/someMetric",
			method: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_200_counter",
			url:    "/value/counter/someMetric",
			method: http.MethodGet,
			want: want{
				code:        http.StatusOK,
				contentType: "plain/text",
			},
		},
		{
			name:   "service_get_404",
			url:    "/value/other/some",
			method: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
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

func Test_list(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Error("Status code mismatch. want:", http.StatusOK, "got:", res.StatusCode)
	}
}

func Test_updateJSON(t *testing.T) {

	tests := testcasesWithBody{
		{
			testcase: testcase{
				name:   "service_post_update_gauge_json_200",
				url:    "/update/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody:  `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
			resBody:  `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
			wantBody: `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			req.Header = map[string][]string{
				"Content-Type": {"application/json"},
			}

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
			// bodyBytes, err := io.ReadAll(res.Body)
			// if err != nil {
			// 	t.Error(err)
			// }
			// var cmp1,cmp2 *mt.Metrics,*mt.Metrics
			// cmp1.UnmarshalJSON(tt.wantBody)

			// for _, v := range cmp1 {

			// if cmp1 != string(bodyBytes) {
			// 	t.Error("Body mismatch. want:", tt.wantBody, "got:", string(bodyBytes))
			// }
		})
	}
}
