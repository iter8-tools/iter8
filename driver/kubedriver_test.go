package driver

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
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKOps(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	kd := NewKubeDriver(cli.New()) // we will ignore this value
	assert.NotNil(t, kd)

	kd = NewFakeKubeDriver(cli.New())
	err := kd.Init()
	assert.NoError(t, err)

	// install
	err = kd.install(action.ChartPathOptions{}, base.CompletePath("../", "charts/iter8"), values.Options{
		Values: []string{"tasks={http}", "http.url=https://httpbin.org/get", "runner=job"},
	}, kd.Group, false)
	assert.NoError(t, err)

	rel, err := kd.Releases.Last(kd.Group)
	assert.NoError(t, err)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.Equal(t, 1, kd.revision)

	err = kd.Init()
	assert.NoError(t, err)

	// upgrade
	err = kd.upgrade(action.ChartPathOptions{}, base.CompletePath("../", "charts/iter8"), values.Options{
		Values: []string{"tasks={http}", "http.url=https://httpbin.org/get", "runner=job"},
	}, kd.Group, false)
	assert.NoError(t, err)

	rel, err = kd.Releases.Last(kd.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 2, rel.Version)
	assert.Equal(t, 2, kd.revision)
	assert.NoError(t, err)

	err = kd.Init()
	assert.NoError(t, err)

	// delete
	err = kd.Delete()
	assert.NoError(t, err)

	// delete
	err = kd.Delete()
	assert.Error(t, err)
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

	kd := NewFakeKubeDriver(cli.New())
	kd.revision = 1

	byteArray, _ := os.ReadFile(ExperimentPath)
	_, _ = kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	// _, _ = kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      "default-1-job",
	// 		Namespace: "default",
	// 		Annotations: map[string]string{
	// 			"iter8.tools/group":    "default",
	// 			"iter8.tools/revision": "1",
	// 		},
	// 	},
	// }, metav1.CreateOptions{})

	err = base.RunExperiment(false, kd)
	assert.NoError(t, err)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
	assert.True(t, metricsServerCalled)

	// check results
	exp, err := base.BuildExperiment(kd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure())
}

func TestLogs(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	kd := NewFakeKubeDriver(cli.New())
	kd.revision = 1

	byteArray, _ := os.ReadFile(base.CompletePath("../testdata/drivertests", ExperimentPath))
	_, _ = kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})
	_, _ = kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job-1831a",
			Namespace: "default",
			Labels: map[string]string{
				"iter8.tools/group": "default",
			},
		},
	}, metav1.CreateOptions{})

	// check logs
	str, err := kd.GetExperimentLogs()
	assert.NoError(t, err)
	assert.Equal(t, "fake logs", str)
}

func TestDryInstall(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	kd := NewFakeKubeDriver(cli.New())

	err := kd.Launch(action.ChartPathOptions{}, base.CompletePath("../", "charts/iter8"), values.Options{
		ValueFiles:   []string{},
		StringValues: []string{},
		Values:       []string{"tasks={http}", "http.url=https://localhost:12345"},
		FileValues:   []string{},
	}, "default", true)

	assert.NoError(t, err)
	assert.FileExists(t, ManifestFile)
}
