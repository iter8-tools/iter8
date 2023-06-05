// Package storageclient provides the storage client for the controllers package
package storageclient

import "github.com/iter8-tools/iter8/base/summarymetrics"

// VersionMetricSummary is a map of metric names to summary metric values
type VersionMetricSummary map[string]summarymetrics.SummaryMetric

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	// GetSummaryMetrics returns versionMetricSummary for each version of the application
	// GetSummaryMetrics is called by Iter8 experiments
	GetSummaryMetrics(applicationName string) (map[int]VersionMetricSummary, error)

	// WriteMetric is called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	// Key 2: kt-metric::my-app::0::my-signature::my-metric::my-user (get the metric value with all the provided information)
	SetMetric(applicationName string, version int, signature, metric, user string, metricValue float64) error

	// Key 3: kt-app-metrics::my-app (get a list of metrics associated with my-app)
	SetMetrics(applicationName, metric string) error
	// GetMetrics(applicationName string) ([]string, error)

	// Key 5: kt-users::my-app::0::my-signature::my-user -> true (get all users for a particular app version)
	SetUsers(applicationName string, version int, signature, user string) error
	// GetUsers(applicationName string, version int, signature) ([]string, error)
}
