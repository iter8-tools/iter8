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

func TestKubeReportText(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))

	byteArray, _ := os.ReadFile(base.CompletePath("../testdata/assertinputs", driver.ExperimentPath))
	_, _ = rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	err := rOpts.KubeRun(os.Stdout)
	assert.NoError(t, err)
}

func TestKubeReportHTML(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.OutputFormat = HTMLOutputFormatKey

	byteArray, _ := os.ReadFile(base.CompletePath("../testdata/assertinputs", driver.ExperimentPath))
	_, _ = rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	err := rOpts.KubeRun(os.Stdout)
	assert.NoError(t, err)
}

func TestKubeReportInvalid(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.OutputFormat = "invalid"

	byteArray, _ := os.ReadFile(base.CompletePath("../testdata/assertinputs", driver.ExperimentPath))
	_, _ = rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	err := rOpts.KubeRun(os.Stdout)
	assert.ErrorContains(t, err, "unsupported report format")
}
