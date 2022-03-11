package action

import (
	"context"
	"io/ioutil"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLocalRun(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	rOpts := NewRunOpts()
	rOpts.RunDir = base.CompletePath("../", "testdata")
	err := rOpts.LocalRun()
	assert.NoError(t, err)

	httpmock.DeactivateAndReset()
}

func TestKubeRun(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	rOpts := NewRunOpts()
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", "experiment.yaml"))
	rOpts.Group = "default"
	rOpts.Revision = 1
	fClientset := fake.NewSimpleClientset()
	fClientset.PrependReactor("create", "secrets", secretDataReactor)
	fClientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"experiment.yaml": byteArray,
		},
	}, metav1.CreateOptions{})
	fClientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})
	rOpts.Clientset = fClientset
	rOpts.EnvSettings = cli.New()
	err := rOpts.KubeRun()
	assert.NoError(t, err)

	// check results
	exp, err := base.BuildExperiment(true, &rOpts.KubeDriver)
	assert.NoError(t, err)
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	assert.True(t, exp.SLOs())
	assert.Equal(t, 4, exp.Result.NumCompletedTasks)

	log.Logger.Info(exp.Result)

	httpmock.DeactivateAndReset()
}
