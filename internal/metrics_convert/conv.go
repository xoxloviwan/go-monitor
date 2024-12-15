package metricsconvert

import (
	api "github.com/xoxloviwan/go-monitor/internal/metrics_types"
	pb "github.com/xoxloviwan/go-monitor/internal/metrics_types/proto"
)

// ConvMetricOne converts an api.Metrics struct to a pb.Metric struct.
// It copies the ID and Type fields, and if the Delta or Value fields
// are not nil, it copies their values to the corresponding fields
// in the pb.Metric struct.
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

// ConvMetricOneInverse converts a pb.Metric struct to an api.Metrics struct.
// It copies the ID and Type fields, and depending on the Type field, it
// copies the Delta or Value field from the pb.Metric struct to the
// corresponding field in the api.Metrics struct.
func ConvMetricOneInverse(m *pb.Metric) *api.Metrics {
	converted := api.Metrics{ID: m.Id, MType: m.Type}
	if m.Type == api.CounterName && m.Delta != 0 {
		converted.Delta = &m.Delta
	} else {
		converted.Value = &m.Value
	}
	return &converted
}

// ConvMetrics converts a slice of api.Metrics structs to a pb.Metrics struct.
// It iterates through the input slice, converting each api.Metrics struct
// to a pb.Metric struct using the ConvMetricOne function, and returns a
// pointer to a pb.Metrics struct containing the converted metrics.
func ConvMetrics(ms []api.Metrics) *pb.Metrics {
	converted := make([]*pb.Metric, len(ms))
	for i := range ms {
		converted[i] = ConvMetricOne(ms[i])
	}
	return &pb.Metrics{Metrics: converted}
}

// ConvMetricsInverse converts a pb.Metrics struct to an api.MetricsList struct.
func ConvMetricsInverse(ms *pb.Metrics) *api.MetricsList {
	converted := make(api.MetricsList, len(ms.Metrics))
	for i := range ms.Metrics {
		converted[i] = *ConvMetricOneInverse(ms.Metrics[i])
	}
	return &converted
}
