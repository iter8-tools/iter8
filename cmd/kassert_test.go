package cmd

import (
	"context"
	"fmt"
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

func TestKAssert(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", base.GetTrackingHandler(&verifyHandlerCalled))

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata", "experiment.tpl"), url, id.ExperimentPath)

	// run test
	testAssert(t, id.ExperimentPath, url, "output/kassert.txt", false)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
}

func TestKAssertFailsSLOs(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", base.GetTrackingHandler(&verifyHandlerCalled))

	// create experiment.yaml
	base.CreateExperimentYaml(t, base.CompletePath("../testdata", "experiment_fails.tpl"), url, id.ExperimentPath)

	// run test
	testAssert(t, id.ExperimentPath, url, "output/kassertfails.txt", true)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
}

func testAssert(t *testing.T, experiment string, url string, expectedOutputFile string, expectError bool) {
	tests := []cmdTestCase{
		// k launch
		{
			name:   "k launch",
			cmd:    fmt.Sprintf("k launch -c %v --localChart --set tasks={http,assess} --set http.url=%s --set http.duration=2s", base.CompletePath("../testdata/charts", "iter8"), url),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// k run
		{
			name: "k run",
			cmd:  "k run -g default --namespace default",
		},
		// k assert
		{
			name:      "k assert",
			cmd:       "k assert -c completed -c nofailure -c slos",
			golden:    base.CompletePath("../testdata", expectedOutputFile),
			wantError: expectError,
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
