// Package storageclient provides the storage client for the controllers package
package storageclient

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	Start()
	CreateMetric(metricName string, value interface{}) error
	GetMetric(metricName string) (interface{}, error)
	DeleteMetric(metricName string) (interface{}, error)
	Size() (int, int, error) // current capacity and max capacity TODO: units?
}
