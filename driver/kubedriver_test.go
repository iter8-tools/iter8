package driver

import (
	"context"
	"os"
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

func TestKOps(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	kd := NewKubeDriver(cli.New()) // we will ignore this value
	assert.NotNil(t, kd)

	kd = NewFakeKubeDriver(cli.New())
	err := kd.Init()
	assert.NoError(t, err)

	// install
	err = kd.install(action.ChartPathOptions{}, base.CompletePath("../", "testdata/charts/iter8"), values.Options{
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
	err = kd.upgrade(action.ChartPathOptions{}, base.CompletePath("../", "testdata/charts/iter8"), values.Options{
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
	_ = os.Chdir(t.TempDir())
	base.SetupWithMock(t)

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

	_, _ = kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
			Annotations: map[string]string{
				"iter8.tools/group":    "default",
				"iter8.tools/revision": "1",
			},
		},
	}, metav1.CreateOptions{})

	err := base.RunExperiment(false, kd)
	assert.NoError(t, err)

	// check results
	exp, err := base.BuildExperiment(kd)
	assert.NoError(t, err)
	assert.True(t, exp.Completed() && exp.NoFailure() && exp.SLOs())
}

func TestLogs(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	base.SetupWithMock(t)

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

	err := kd.Launch(action.ChartPathOptions{}, base.CompletePath("../", "testdata/charts/iter8"), values.Options{
		ValueFiles:   []string{},
		StringValues: []string{},
		Values:       []string{"tasks={http}", "http.url=https://localhost:12345"},
		FileValues:   []string{},
	}, "default", true)

	assert.NoError(t, err)
	assert.FileExists(t, ManifestFile)
}
