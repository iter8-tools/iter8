package base

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

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
