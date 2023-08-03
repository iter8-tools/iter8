package cmd

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
	id "github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKRun(t *testing.T) {
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

			fmt.Println(string(body))

			if _, ok := bodyFortioResult.EndpointResults[url]; !ok {
				assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain call: %s", url))
			}
		},
	})

	_ = os.Chdir(t.TempDir())

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata", "experiment.tpl"), url, id.ExperimentPath)

	tests := []cmdTestCase{
		// k report
		{
			name:   "k run",
			cmd:    "k run -g default --namespace default",
			golden: base.CompletePath("../testdata", "output/krun.txt"),
		},
	}

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)

	// and read it...
	byteArray, _ := os.ReadFile(id.ExperimentPath)
	_, _ = kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
	assert.True(t, metricsServerCalled)
}
