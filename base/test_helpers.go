package base

import (
	"testing"
	"text/template"

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

	metricsTemplate *template.Template
}

// ReadResult enables results to be read from the mock driver
func (m *mockDriver) ReadResult() (*ExperimentResult, error) {
	return m.Experiment.Result, nil
}

// ReadMetricsSpec enables metrics spec to be read from the mock driver
func (m *mockDriver) ReadMetricsSpec(provider string) (*template.Template, error) {
	return m.metricsTemplate, nil
}

// WriteResult enables results to be written from the mock driver
func (m *mockDriver) WriteResult(r *ExperimentResult) error {
	m.Experiment.Result = r
	return nil
}

// ReadSpec enables spec to be read from the mock secret
func (m *mockDriver) ReadSpec() (ExperimentSpec, error) {
	return m.Experiment.Tasks, nil
}
