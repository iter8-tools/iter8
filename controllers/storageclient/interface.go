// Package storageclient provides the storage client for the controllers package
package storageclient

import "github.com/iter8-tools/iter8/base/summarymetrics"

// versionMetricSummary is a map of metric names to summary metric values
type versionMetricSummary map[string]summarymetrics.SummaryMetric

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	// WriteMetric is called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	WriteMetric(applicationName string, version string, user string, metricName string, metricValue float64) error

	// GetSummaryMetrics returns versionMetricSummary for each version of the application
	// GetSummaryMetrics is called by Iter8 experiments
	GetSummaryMetrics(applicationName string) (map[int]versionMetricSummary, error)

	// Key 1: kt-signature::my-app::0 (get the signature of the last version)
	GetSignature(applicationName string, version int) (string, error)
	SetSignature(applicationName string, version int, signature string) error

	// Key 2: kt-metric::my-app::0::my-signature::my-metric::my-user (get the metric value with all the provided information)
	GetMetric(applicationName string, version int, signature, metric, user string) (float64, error)
	SetMetric(applicationName string, version int, signature, metric, user string, metricValue float64) error

	// Key 3: kt-app-metrics::my-app (get a list of metrics associated with my-app)
	GetMetrics(applicationName string) ([]string, error)

	// Key 4: kt-app-versions::my-app (get a number of versions for my-app)
	GetVersions(applicationName string) (int, error)

	// Key 5: kt-users::my-app::0::my-signature::my-user -> true (get all users for a particular app version) (getDistinctUserCt())
	GetUsers(applicationName string, version int, signature, user string) ([]string, error)
}
