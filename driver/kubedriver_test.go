package driver

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubeRun(t *testing.T) {
	base.SetupWithMock(t)

	kd := NewFakeKubeDriver(cli.New())
	kd.Revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../", "testdata/drivertests/experiment.yaml"))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"experiment.yaml": byteArray,
		},
	}, metav1.CreateOptions{})
	kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	err := base.RunExperiment(kd)
	assert.NoError(t, err)

	// check results
	exp, err := base.BuildExperiment(true, kd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}
