package base

import (
	"errors"
	"io"
	"os"
	"path/filepath"
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

// CopyFileToPwd
func CopyFileToPwd(t *testing.T, filePath string) error {
	// get file
	srcFile, err := os.Open(filePath)
	if err != nil {
		return errors.New("could not open metrics file")
	}
	t.Cleanup(func() { srcFile.Close() })

	// create copy of file in pwd
	destFile, err := os.Create(filepath.Base(filePath))
	if err != nil {
		return errors.New("could not create copy of metrics file in temp directory")
	}
	t.Cleanup(func() {
		destFile.Close()
	})
	io.Copy(destFile, srcFile)
	return nil
}
