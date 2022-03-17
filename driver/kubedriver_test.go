package driver

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelm(t *testing.T) {
	srv := SetupWithRepo(t)
	kd := NewKubeDriver(cli.New()) // we will ignore this value
	assert.NotNil(t, kd)
	kd = NewFakeKubeDriver(cli.New())
	err := kd.Init()
	assert.NoError(t, err)

	// install
	err = kd.Install(">=0.0.0", "load-test-http", values.Options{
		Values: []string{"url=https://httpbin.org/get"},
	}, kd.Group, false, &action.ChartPathOptions{
		RepoURL: srv.URL(),
	})
	assert.NoError(t, err)

	rel, err := kd.Releases.Last(kd.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.Equal(t, 1, kd.Revision)
	assert.NoError(t, err)

	err = kd.Init()
	assert.NoError(t, err)

	// upgrade
	err = kd.Upgrade(">=0.0.0", "load-test-http", values.Options{
		Values: []string{"url=https://httpbin.org/get"},
	}, kd.Group, false, &action.ChartPathOptions{
		RepoURL: srv.URL(),
	})
	assert.NoError(t, err)

	rel, err = kd.Releases.Last(kd.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 2, rel.Version)
	assert.Equal(t, 2, kd.Revision)
	assert.NoError(t, err)

	err = kd.Init()
	assert.NoError(t, err)

}

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
		StringData: map[string]string{"experiment.yaml": string(byteArray)},
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
	kd.Revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../", "testdata/drivertests/experiment.yaml"))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		StringData: map[string]string{"experiment.yaml": string(byteArray)},
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
