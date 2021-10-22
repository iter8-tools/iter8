package core

import "strings"

// Backend defines a metrics backend
type Backend struct {
	// Name is name of the backend
	Name string `json:"name" yaml:"name"`

	// Text description of the backend
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Provider identifies the type of metrics backend. Used for informational purposes.
	Provider *string `json:"provider,omitempty" yaml:"provider,omitempty"`

	// Metrics is list of metrics available from this backend
	Metrics []Metric `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

// MetricType identifies the type of the metric.
type MetricType string

const (
	// CounterMetricType corresponds to Prometheus Counter metric type
	CounterMetricType MetricType = "Counter"

	// GaugeMetricType is an enhancement of Prometheus Gauge metric type
	GaugeMetricType MetricType = "Gauge"
)

// Metric defines a metric
type Metric struct {
	// Name of the metric
	Name string `json:"name" yaml:"name"`

	// Text description of the metric
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Units of the metric. Used for informational purposes.
	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

	// Type of the metric
	Type MetricType `json:"type" yaml:"type"`
}

var IFBackend Backend

func init() {
	// initialize the backend
	IFBackend = Backend{
		Name:        "iter8-fortio",
		Description: StringPointer("Iter8's built-in backend that supplies latency and error metrics"),
		Provider:    StringPointer("Iter8"),
		Metrics:     []Metric{},
	}

	ifb := &IFBackend

	// start adding metrics to this backend
	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "request-count",
		Description: StringPointer("number of requests sent per-version"),
		Units:       nil,
		Type:        CounterMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "error-count",
		Description: StringPointer("number of error responses"),
		Units:       nil,
		Type:        CounterMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "error-rate",
		Description: StringPointer("fraction of requests with error responses"),
		Units:       nil,
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "mean-latency",
		Description: StringPointer("average request latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "min-latency",
		Description: StringPointer("minimum request latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "max-latency",
		Description: StringPointer("maximum request latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "stddev-latency",
		Description: StringPointer("standard deviation of request latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "p50",
		Description: StringPointer("50th percentile (median) latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "p75",
		Description: StringPointer("75th percentile (tail) latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "p90",
		Description: StringPointer("90th percentile (tail) latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "p99",
		Description: StringPointer("99th percentile (tail) latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

	ifb.Metrics = append(ifb.Metrics, Metric{
		Name:        "p99.9",
		Description: StringPointer("99.9th percentile (tail) latency"),
		Units:       StringPointer("sec"),
		Type:        GaugeMetricType,
	})

}

// HasMetric checks if the backend contains a metric
func (b *Backend) HasMetric(m string) bool {
	if strings.HasPrefix(m, b.Name) {
		for i := range b.Metrics {
			if b.Name+"/"+b.Metrics[i].Name == m {
				return true
			}
		}
	}
	return false
}
