package driver

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestKOps(t *testing.T) {
	kd := NewKubeDriver(cli.New()) // we will ignore this value
	assert.NotNil(t, kd)

	kd = NewFakeKubeDriver(cli.New())
	err := kd.Init()
	assert.NoError(t, err)

	// install
	err = kd.install(base.CompletePath("../", "charts/load-test-http"), values.Options{
		Values: []string{"url=https://httpbin.org/get"},
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
	err = kd.upgrade(base.CompletePath("../", "charts/load-test-http"), values.Options{
		Values: []string{"url=https://httpbin.org/get"},
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
	base.SetupWithMock(t)

	kd := NewFakeKubeDriver(cli.New())
	kd.revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/drivertests", ExperimentSpecPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{ExperimentSpecPath: string(byteArray)},
	}, metav1.CreateOptions{})

	resultBytes, _ := yaml.Marshal(base.ExperimentResult{
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
		Iter8Version:      base.MajorMinor,
	})
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-result",
			Namespace: "default",
		},
		StringData: map[string]string{ExperimentResultPath: string(resultBytes)},
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

func TestLogs(t *testing.T) {
	base.SetupWithMock(t)

	kd := NewFakeKubeDriver(cli.New())
	kd.revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/drivertests", ExperimentSpecPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{ExperimentSpecPath: string(byteArray)},
	}, metav1.CreateOptions{})
	kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job-1831a",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "default-1-job",
			},
		},
	}, metav1.CreateOptions{})

	// check logs
	str, err := kd.GetExperimentLogs()
	assert.NoError(t, err)
	assert.Equal(t, "fake logs", str)
}
