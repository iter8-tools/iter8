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
}
