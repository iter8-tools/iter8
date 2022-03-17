package action

import (
	"context"
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLog(t *testing.T) {
	srv := SetupWithRepo(t)

	var err error

	// fix launchOpts
	launchOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	launchOpts.ChartName = "load-test-http"
	launchOpts.DestDir = t.TempDir()
	launchOpts.Values = []string{"url=https://iter8.tools", "duration=2s"}
	launchOpts.RepoURL = srv.URL()

	err = launchOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := launchOpts.Releases.Last(launchOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)

	// fix lOpts
	lOpts := NewLogOpts(launchOpts.KubeDriver)
	lOpts.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job-8218s",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "default-1-job",
			},
		},
	}, metav1.CreateOptions{})

	str, err := lOpts.KubeRun()
	assert.NoError(t, err)
	assert.Equal(t, "fake logs", str)
}
