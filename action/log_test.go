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

func TestLog(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = base.CompletePath("../", "iter8")
	lOpts.LocalChart = true
	lOpts.Values = []string{"tasks={http}", "http.url=https://httpbin.org/get", "http.duration=2s"}

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.Equal(t, 1, rel.Version)
	assert.NoError(t, err)

	// fix lOpts
	logOpts := NewLogOpts(lOpts.KubeDriver)
	_, _ = logOpts.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job-8218s",
			Namespace: "default",
			Labels: map[string]string{
				"iter8.tools/group": "default",
			},
		},
	}, metav1.CreateOptions{})

	str, err := logOpts.KubeRun()
	assert.NoError(t, err)
	assert.Equal(t, "fake logs", str)
}
