package base

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

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

// CreateExperimentYaml creates an experiment.yaml file from a template and a URL
func CreateExperimentYaml(t *testing.T, template string, url string, output string) {
	values := struct {
		URL string
	}{
		URL: url,
	}

	byteArray, err := os.ReadFile(filepath.Clean(template))
	assert.NoError(t, err)

	tpl, err := CreateTemplate(string(byteArray))
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = tpl.Execute(&buf, values)
	assert.NoError(t, err)

	err = os.WriteFile(output, buf.Bytes(), 0600)
	assert.NoError(t, err)
}

// GetTrackingHandler creates a handler for fhttp.DynamicHTTPServer that sets a variable to true
// This can be used to verify that the handler was called.
func GetTrackingHandler(breadcrumb *bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		*breadcrumb = true
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}
}

// StartHTTPMock activates and cleanups httpmock
func StartHTTPMock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
}

// MetricsServerCallback is a callback function for when the particular metrics server endpoint
// is called
type MetricsServerCallback func(req *http.Request)

// MockMetricsServerInput is the input for MockMetricsServer()
// allows the user to provide callbacks when particular endpoints are called
type MockMetricsServerInput struct {
	MetricsServerURL string

	// PUT /experimentResult
	ExperimentResultCallback MetricsServerCallback
	// GET /grpcDashboard
	GRPCDashboardCallback MetricsServerCallback
	// GET /httpDashboard
	HTTPDashboardCallback MetricsServerCallback
}

// MockMetricsServer is a mock metrics server
// use the callback functions in the MockMetricsServerInput to test if those endpoints are called
func MockMetricsServer(input MockMetricsServerInput) {
	// PUT /experimentResult
	httpmock.RegisterResponder(
		http.MethodPut,
		input.MetricsServerURL+ExperimentResultPath,
		func(req *http.Request) (*http.Response, error) {
			if input.ExperimentResultCallback != nil {
				input.ExperimentResultCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// GET /httpDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.MetricsServerURL+HTTPDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.HTTPDashboardCallback != nil {
				input.HTTPDashboardCallback(req)
			}

			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// GET /grpcDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.MetricsServerURL+GRPCDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.GRPCDashboardCallback != nil {
				input.GRPCDashboardCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)
}
