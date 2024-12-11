package metricsconvert

import (
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"
)

func ConvMetricOne(m api.Metrics) *pb.Metric {
	converted := pb.Metric{Id: m.ID, Type: m.MType}
	if m.Delta != nil {
		converted.Delta = *m.Delta
	}
	if m.Value != nil {
		converted.Value = *m.Value
	}
	return &converted
}

func ConvMetricOneInverse(m *pb.Metric) *api.Metrics {
	converted := api.Metrics{ID: m.Id, MType: m.Type}
	if m.Type == api.CounterName && m.Delta != 0 {
		converted.Delta = &m.Delta
	} else {
		converted.Value = &m.Value
	}
	return &converted
}

func ConvMetrics(ms []api.Metrics) *pb.Metrics {
	converted := make([]*pb.Metric, len(ms))
	for i := range ms {
		converted[i] = ConvMetricOne(ms[i])
	}
	return &pb.Metrics{Metrics: converted}
}

func ConvMetricsInverse(ms *pb.Metrics) *api.MetricsList {
	converted := make(api.MetricsList, len(ms.Metrics))
	for i := range ms.Metrics {
		converted[i] = *ConvMetricOneInverse(ms.Metrics[i])
	}
	return &converted
}
