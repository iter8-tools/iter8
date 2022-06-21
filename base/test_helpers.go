package base

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

// SetupWithMock mocks an HTTP endpoint and registers and cleanup function
func SetupWithMock(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))
	t.Cleanup(httpmock.DeactivateAndReset)
}

// mockDriver is a mock driver that can be used to run experiments
type mockDriver struct {
	*Experiment
}

// Read an experiment
func (m *mockDriver) Read() (*Experiment, error) {
	return m.Experiment, nil
}

// Write an experiment
func (m *mockDriver) Write(e *Experiment) error {
	m.Experiment = e
	return nil
}

// GetRevision gets experiment revision
func (m *mockDriver) GetRevision() int {
	return 0
}
