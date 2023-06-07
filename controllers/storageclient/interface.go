// Package storageclient provides the storage client for the controllers package
package storageclient

// SummarizedMetric is a summarization
type SummarizedMetric struct {
	Count  uint
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

// MetricSummary
type MetricSummary struct {
	SummaryOverUsers        SummarizedMetric
	SummaryOverTransactions SummarizedMetric
}

// VersionMetricSummary is a summarization of metrics for a given version
type VersionMetricSummary struct {
	NumUsers uint64

	// key = metric name; value is the summary value of the metric
	MetricSummaries map[string]MetricSummary
}

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	GetSummaryMetrics(applicationName string, version int, signature string) (*VersionMetricSummary, error)

	// called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	// Example key: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> my-metric-value (get the metric value with all the provided information)
	SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error

	// Example key: kt-users::my-app::0::my-signature::my-user -> true
	SetUser(applicationName string, version int, signature, user string) error
}
