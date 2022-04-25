package action

import (
	"context"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLog(t *testing.T) {
	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartsParentDir = base.CompletePath("../", "")
	lOpts.ChartName = "load-test-http"
	lOpts.NoDownload = true
	lOpts.Values = []string{"url=https://httpbin.org/get", "duration=2s"}
	lOpts.RunDir = t.TempDir()

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)

	// fix lOpts
	logOpts := NewLogOpts(lOpts.KubeDriver)
	logOpts.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-job-8218s",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "default-job",
			},
		},
	}, metav1.CreateOptions{})

	str, err := logOpts.KubeRun()
	assert.NoError(t, err)
	assert.Equal(t, "fake logs", str)
}
