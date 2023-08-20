package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	myName      = "myName"
	myNamespace = "myNamespace"
)

func TestKAssert(t *testing.T) {
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
		ExperimentResultCallback: func(req *http.Request) {
			metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("experiment"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyExperimentResult := base.ExperimentResult{}
			err = json.Unmarshal(body, &bodyExperimentResult)
			assert.NoError(t, err)
			assert.NotNil(t, body)
		},
	})

	_ = os.Chdir(t.TempDir())

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata", "experiment.tpl"), url, id.ExperimentPath)

	// run test
	testAssert(t, id.ExperimentPath, url, "output/kassert.txt", false)
	assert.True(t, metricsServerCalled)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
}

func testAssert(t *testing.T, experiment string, url string, expectedOutputFile string, expectError bool) {
	tests := []cmdTestCase{
		// k launch
		{
			name:   "k launch",
			cmd:    fmt.Sprintf("k launch -c %v --localChart --set tasks={http} --set http.url=%s --set http.duration=2s", base.CompletePath("../charts", "iter8"), url),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// k run
		{
			name: "k run",
			cmd:  "k run -g default --namespace default",
		},
	}

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)

	// read experiment from file created by caller
	byteArray, _ := os.ReadFile(filepath.Clean(experiment))
	_, _ = kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	_, _ = kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
