package base

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

func SetupWithMock(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))
	t.Cleanup(httpmock.DeactivateAndReset)
}

type mockDriver struct {
	*Experiment
}

func (m *mockDriver) ReadResult() (*ExperimentResult, error) {
	return m.Experiment.Result, nil
}

func (m *mockDriver) WriteResult(r *ExperimentResult) error {
	m.Experiment.Result = r
	return nil
}

func (m *mockDriver) ReadSpec() (ExperimentSpec, error) {
	return m.Experiment.Tasks, nil
}
