package base

// HistBucket is a single bucket in a histogram
type HistBucket struct {
	// Lower endpoint of a histogram bucket
	Lower float64 `json:"lower" yaml:"lower"`
	// Upper endpoint of a histogram bucket
	Upper float64 `json:"upper" yaml:"upper"`
	// Count is the frequency count of the bucket
	Count uint64 `json:"count" yaml:"count"`
}

// MetricType identifies the type of the metric.
type MetricType string

// AggregationType identifies the type of the metric aggregator.
type AggregationType string

const (
	// CounterMetricType corresponds to Prometheus Counter metric type
	CounterMetricType MetricType = "Counter"
	// GaugeMetricType corresponds to Prometheus Gauge metric type
	GaugeMetricType MetricType = "Gauge"
	// HistogramMetricType corresponds to a Histogram metric type
	HistogramMetricType MetricType = "Histogram"
	// SampleMetricType corresponds to a Sample metric type
	SampleMetricType MetricType = "Sample"

	// decimalRegex is the regex used to identify percentiles
	decimalRegex = `^([\d]+(\.[\d]*)?|\.[\d]+)$`

	// CountAggregator corresponds to aggregation of type count
	CountAggregator AggregationType = "count"
	// MeanAggregator corresponds to aggregation of type mean
	MeanAggregator AggregationType = "mean"
	// StdDevAggregator corresponds to aggregation of type stddev
	StdDevAggregator AggregationType = "stddev"
	// MinAggregator corresponds to aggregation of type min
	MinAggregator AggregationType = "min"
	// MaxAggregator corresponds to aggregation of type max
	MaxAggregator AggregationType = "max"
	// PercentileAggregator corresponds to aggregation of type max
	PercentileAggregator AggregationType = "percentile"
	// PercentileAggregatorPrefix corresponds to prefix for percentiles
	PercentileAggregatorPrefix = "p"
)
