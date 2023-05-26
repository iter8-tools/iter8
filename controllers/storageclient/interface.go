// Package storageclient provides the storage client for the controllers package
package storageclient

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	// CreateMetric(name string, value interface{}) error
	// GetMetric(name string) (interface{}, error)
	// DeleteMetric(name string) (interface{}, error)
}
