package action

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/time"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestLocalRun(t *testing.T) {
	base.SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.RunDir = base.CompletePath("../", "testdata")
	err := rOpts.LocalRun()
	assert.NoError(t, err)
}

func TestKubeRun(t *testing.T) {
	base.SetupWithMock(t)
	// fix rOpts
	rOpts := NewRunOpts(driver.NewFakeKubeDriver(cli.New()))

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", driver.ExperimentSpecPath))
	rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentSpecPath: string(byteArray)},
	}, metav1.CreateOptions{})

	resultBytes, _ := yaml.Marshal(base.ExperimentResult{
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
		Iter8Version:      base.MajorMinor,
	})
	rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-result",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentResultPath: string(resultBytes)},
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
