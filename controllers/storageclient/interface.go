// Package storageclient provides the storage client for the controllers package
package storageclient

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	WriteMetric(applicationName string, version string, user string, metricName string, metricValue float64) error
	GetMetric(applicationName string, version string, user string, metricName string) (float64, error)
	DeleteMetric(applicationName string, version string, user string, metricName string) (float64, error)

	FreeSpace(bytes uint64) error
}
