// Package storageclient provides the storage client for the controllers package
package storageclient

// SummarizedMetric is a metric summary
type SummarizedMetric struct {
	Count  uint64
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

// MetricSummary contains metric summary for all metrics as well as cumulative metrics per user
type MetricSummary struct {
	// all transactions
	SummaryOverTransactions SummarizedMetric

	// cumulative metrics per user
	SummaryOverUsers SummarizedMetric
}

// VersionMetricSummary is a metric summary for a given app version
type VersionMetricSummary struct {
	NumUsers uint64

	// key = metric name; value is the metric summary
	MetricSummaries map[string]MetricSummary
}

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	GetSummaryMetrics(applicationName string, version int, signature string) (*VersionMetricSummary, error)

	// Returns a nested map of the metrics data for a particular application, user, and transaction
	// Example:
	//	{
	//		"my-metric": {
	//			"my-user": {
	//				"my-transaction-id": 5.0
	//			}
	//		}
	//	}
	GetMetrics(applicationName string, version int, signature string) (map[string]map[string]map[string]float64, error)

	// called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	// Example key: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> my-metric-value (get the metric value with all the provided information)
	SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error

	// Example key: kt-users::my-app::0::my-signature::my-user -> true
	SetUser(applicationName string, version int, signature, user string) error
}
