// Package storageclient provides the storage client for the controllers package
package storageclient

import "github.com/iter8-tools/iter8/base/summarymetrics"

// VersionMetricSummary is a map of metric names to summary metric values
type VersionMetricSummary map[string]summarymetrics.SummaryMetric

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	// GetSummaryMetrics returns versionMetricSummary for each version of the application
	// called by Iter8 experiments
	GetSummaryMetrics(applicationName string) (*map[int]VersionMetricSummary, error)

	// called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	// Example key: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> my-metric-value (get the metric value with all the provided information)
	SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error

	// Example key: kt-users::my-app::0::my-signature::my-user -> true
	SetUser(applicationName string, version int, signature, user string) error
}
