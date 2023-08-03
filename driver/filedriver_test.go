package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	myName      = "myName"
	myNamespace = "myNamespace"
)

func startHTTPMock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
}

func TestLocalRun(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(base.MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", base.GetTrackingHandler(&verifyHandlerCalled))

	// mock metrics server
	base.StartHTTPMock(t)
	metricsServerCalled := false
	base.MockMetricsServer(base.MockMetricsServerInput{
		MetricsServerURL: metricsServerURL,
		PerformanceResultCallback: func(req *http.Request) {
			metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("experiment"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyFortioResult := base.FortioResult{}
			err = json.Unmarshal(body, &bodyFortioResult)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			if _, ok := bodyFortioResult.EndpointResults[url]; !ok {
				assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain url: %s", url))
			}
		},
	})

	_ = os.Chdir(t.TempDir())

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata/drivertests", "experiment.tpl"), url, ExperimentPath)

	fd := FileDriver{
		RunDir: ".",
	}
	err = base.RunExperiment(false, &fd)
	assert.NoError(t, err)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)

	// check results
	exp, err := base.BuildExperiment(&fd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure())
	assert.True(t, metricsServerCalled)
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
