package action

import (
	"context"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLocalRun(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	_ = driver.CopyFileToPwd(t, base.CompletePath("../", "testdata/experiment.yaml"))

	base.SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))
	err := rOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeRun(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	base.SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))

	byteArray, _ := os.ReadFile(base.CompletePath("../testdata", driver.ExperimentPath))
	_, _ = rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	err := rOpts.KubeRun()
	assert.NoError(t, err)

	// check results
	exp, err := base.BuildExperiment(rOpts.KubeDriver)
	assert.NoError(t, err)
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	assert.True(t, exp.SLOs())
	assert.Equal(t, 4, exp.Result.NumCompletedTasks)
}
