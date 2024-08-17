package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	mock "github.com/xoxloviwan/go-monitor/internal/api/mock"
	mt "github.com/xoxloviwan/go-monitor/internal/metrics_types"
)

type want struct {
	code        int
	contentType string
	err         error
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
	reqBody          string
	wantBody         string
	lastCounterValue int64
	lastGaugeValue   float64
}

func setup(t *testing.T) (*gin.Engine, *mock.MockReaderWriter) {
	ping := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewMockReaderWriter(ctrl)
	gin.SetMode(gin.ReleaseMode)
	return SetupRouter(ping, m, slog.LevelError), m
}

func Test_update_value(t *testing.T) {

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
				err:         errors.New("unknown metric type"),
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
	router, m := setup(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			urlSpl := strings.Split(tt.url, "/")
			var metricType string
			var metricName string
			var metricValue string
			if len(urlSpl) > 2 {
				req.SetPathValue("metricType", urlSpl[2])
				metricType = urlSpl[2]
			}
			if len(urlSpl) > 3 {
				req.SetPathValue("metricName", urlSpl[3])
				metricName = urlSpl[3]
			}

			if len(urlSpl) > 4 {
				req.SetPathValue("metricValue", urlSpl[4])
				metricValue = urlSpl[4]
			}
			m.EXPECT().Add(metricType, metricName, metricValue).Return(tt.want.err)
			m.EXPECT().Get(metricType, metricName).Return(gomock.Any().String(), tt.want.err == nil)

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

	router, m := setup(t)
	m.EXPECT().String().Return("some string")
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
			reqBody:          `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
			wantBody:         `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
			lastCounterValue: 0,
			lastGaugeValue:   23.4,
		},
		{
			testcase: testcase{
				name:   "service_post_update_counter_json_200_1",
				url:    "/update/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody:          `{"id": "someMetric", "type": "counter", "delta": 23}`,
			wantBody:         `{"id": "someMetric", "type": "counter", "delta": 23}`,
			lastCounterValue: 0,
			lastGaugeValue:   0,
		},
		{
			testcase: testcase{
				name:   "service_post_update_counter_json_200_2",
				url:    "/update/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody:          `{"id": "someMetric", "type": "counter", "delta": 20}`,
			wantBody:         `{"id": "someMetric", "type": "counter", "delta": 43}`, // сохранилось значение от теста service_post_update_counter_json_200
			lastCounterValue: 23,
			lastGaugeValue:   0,
		},
	}

	router, m := setup(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var err error

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			req.Header = map[string][]string{
				"Content-Type": {"application/json"},
			}

			gotInput := mt.Metrics{}
			if err = gotInput.UnmarshalJSON([]byte(tt.reqBody)); err != nil {
				t.Error(err)
			}
			gotInputList := &mt.MetricsList{gotInput}
			m.EXPECT().AddMetrics(gomock.Any(), gotInputList).Return(nil).Times(1)

			gotOutputList := &mt.MetricsList{gotInput}
			m.EXPECT().GetMetrics(gomock.Any(), gotOutputList).DoAndReturn(func(ctx context.Context, gotOutputList *mt.MetricsList) error {
				wantOutputList := *gotOutputList
				if wantOutputList[0].MType == "counter" {
					val := tt.lastCounterValue
					val += *wantOutputList[0].Delta
					wantOutputList[0].Delta = new(int64)
					wantOutputList[0].Delta = &val
				}
				return nil
			}).Times(1)

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
			var bodyBytes []byte
			bodyBytes, err = io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			var got = mt.Metrics{}
			var want = mt.Metrics{}
			if err = got.UnmarshalJSON(bodyBytes); err != nil {
				t.Error(err)
			}
			if err = want.UnmarshalJSON([]byte(tt.wantBody)); err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_valueJSON(t *testing.T) {

	tests := testcasesWithBody{
		{
			testcase: testcase{
				name:   "service_post_value_gauge_json_200",
				url:    "/value/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody:          `{"id": "someMetric", "type": "gauge"}`,
			wantBody:         `{"id": "someMetric", "type": "gauge", "value": 23.4}`,
			lastCounterValue: 0,
			lastGaugeValue:   23.4,
		},
		{
			testcase: testcase{
				name:   "service_post_value_counter_json_200",
				url:    "/value/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody:          `{"id": "someMetric", "type": "counter"}`,
			wantBody:         `{"id": "someMetric", "type": "counter", "delta": 43}`, // сохранилось значение от теста service_post_update_counter_json_200
			lastCounterValue: 43,
			lastGaugeValue:   0,
		},
	}

	router, m := setup(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var err error

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			req.Header = map[string][]string{
				"Content-Type": {"application/json"},
			}

			gotInput := mt.Metrics{}
			if err = gotInput.UnmarshalJSON([]byte(tt.reqBody)); err != nil {
				t.Error(err)
			}

			m.EXPECT().Get(gotInput.MType, gotInput.ID).DoAndReturn(func(metricType string, metricName string) (string, bool) {
				fmt.Printf("%s last counter value: %d\n", tt.name, tt.lastCounterValue)
				if metricType == "counter" {
					return fmt.Sprintf("%d", tt.lastCounterValue), true
				} else {
					return fmt.Sprintf("%f", tt.lastGaugeValue), true
				}
			}).Times(1)

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
			var bodyBytes []byte
			bodyBytes, err = io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			var got = mt.Metrics{}
			var want = mt.Metrics{}
			if err = got.UnmarshalJSON(bodyBytes); err != nil {
				t.Error(err)
			}
			if err = want.UnmarshalJSON([]byte(tt.wantBody)); err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_updatesJSON(t *testing.T) {

	tests := testcasesWithBody{
		{
			testcase: testcase{
				name:   "service_post_updates_counter_json_200",
				url:    "/updates/",
				method: http.MethodPost,
				want: want{
					code:        http.StatusOK,
					contentType: "application/json",
				},
			},
			reqBody: `[
				{"id":"someMetric","type":"counter","delta":0},
				{"id":"someMetric","type":"counter","delta":10},
				{"id":"someMetric","type":"counter","delta":20}
			]`,
			wantBody: `[{"id": "someMetric", "type": "counter", "delta": 30}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var err error

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			req.Header = map[string][]string{
				"Content-Type": {"application/json"},
			}

			router, m := setup(t)

			gotInputList := &mt.MetricsList{}
			if err = gotInputList.UnmarshalJSON([]byte(tt.reqBody)); err != nil {
				t.Error(err)
			}

			m.EXPECT().AddMetrics(gomock.Any(), gotInputList).Return(nil).Times(1)

			m.EXPECT().GetMetrics(gomock.Any(), gotInputList).DoAndReturn(func(ctx context.Context, gotOutputList *mt.MetricsList) error {
				*gotOutputList = mt.MetricsList{}
				res := mt.Metrics{
					ID:    "someMetric",
					MType: "counter",
				}
				var val int64 = 30
				res.Delta = new(int64)
				res.Delta = &val
				*gotOutputList = append(*gotOutputList, res)
				return nil
			}).Times(1)

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.want.code != res.StatusCode {
				t.Error("Status code mismatch. want:", tt.want.code, "got:", res.StatusCode)
			}
			var bodyBytes []byte
			bodyBytes, err = io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			var got = mt.MetricsList{}
			var want = mt.MetricsList{}
			if err = got.UnmarshalJSON(bodyBytes); err != nil {
				t.Error(err)
			}
			if err = want.UnmarshalJSON([]byte(tt.wantBody)); err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
