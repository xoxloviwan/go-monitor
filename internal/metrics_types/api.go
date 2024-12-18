package metrictypes

//go:generate easyjson -output_filename api_easyjson_generated.go -all api.go

//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

//easyjson:json
type MetricsList []Metrics

// CounterName is a constant representing the counter metric type.
const CounterName = "counter"

// GaugeName is a constant representing the gauge metric type.
const GaugeName = "gauge"
