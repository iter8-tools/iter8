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
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestLocalRun(t *testing.T) {
	SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.RunDir = base.CompletePath("../", "testdata")
	err := rOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeRun(t *testing.T) {
	SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.Revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", "experiment.yaml"))
	rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"experiment.yaml": byteArray,
		},
	}, metav1.CreateOptions{})
	rOpts.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})
	err := rOpts.KubeRun()
	assert.NoError(t, err)

	// check results
	exp, err := base.BuildExperiment(true, rOpts.KubeDriver)
	assert.NoError(t, err)
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	assert.True(t, exp.SLOs())
	assert.Equal(t, 4, exp.Result.NumCompletedTasks)

	log.Logger.Debug(exp.Result)
}
