package driver

import (
	"fmt"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
)

func TestLocalRun(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", base.GetTrackingHandler(&verifyHandlerCalled))

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata/drivertests", "experiment.tpl"), url, ExperimentPath)

	fd := FileDriver{
		RunDir: ".",
	}
	err := base.RunExperiment(false, &fd)
	assert.NoError(t, err)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)

	// check results
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}

func TestFileDriverReadError(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	fd := FileDriver{
		RunDir: ".",
	}
	exp, err := fd.Read()
	assert.Error(t, err)
	assert.Nil(t, exp)
}
