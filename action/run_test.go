package action

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	myName      = "myName"
	myNamespace = "myNamespace"
)

// TODO: duplicated from collect_http_test.go
func startHTTPMock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
}

type DashboardCallback func(req *http.Request)

type mockMetricsServerInput struct {
	metricsServerURL string

	// GET /httpDashboard
	httpDashboardCallback DashboardCallback
	// GET /grpcDashboard
	gRPCDashboardCallback DashboardCallback
	// PUT /performanceResult
	performanceResultCallback DashboardCallback
}

func mockMetricsServer(input mockMetricsServerInput) {
	// GET /httpDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.metricsServerURL+base.HTTPDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.httpDashboardCallback != nil {
				input.httpDashboardCallback(req)
			}

			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// GET /grpcDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.metricsServerURL+base.GRPCDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.gRPCDashboardCallback != nil {
				input.gRPCDashboardCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// PUT /performanceResult
	httpmock.RegisterResponder(
		http.MethodPut,
		input.metricsServerURL+base.PerformanceResultPath,
		func(req *http.Request) (*http.Response, error) {
			if input.performanceResultCallback != nil {
				input.performanceResultCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)
}

func TestKubeRun(t *testing.T) {
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
	startHTTPMock(t)
	metricsServerCalled := false
	mockMetricsServer(mockMetricsServerInput{
		metricsServerURL: metricsServerURL,
		performanceResultCallback: func(req *http.Request) {
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

			// if _, ok := bodyFortioResult.EndpointResults[call]; !ok {
			// 	assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain call: %s", call))
			// }
		},
	})

	_ = os.Chdir(t.TempDir())

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata", "experiment.tpl"), url, driver.ExperimentPath)

	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))

	// read experiment from file created above
	byteArray, _ := os.ReadFile(driver.ExperimentPath)
	_, _ = rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	err = rOpts.KubeRun()
	assert.NoError(t, err)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
	assert.True(t, metricsServerCalled)

	// check results
	exp, err := base.BuildExperiment(rOpts.KubeDriver)
	assert.NoError(t, err)
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	assert.Equal(t, 1, exp.Result.NumCompletedTasks)

}
