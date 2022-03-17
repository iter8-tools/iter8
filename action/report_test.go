package action

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLocalReportText(t *testing.T) {
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.RunDir = base.CompletePath("../", "testdata/assertinputs")

	err := rOpts.LocalRun(os.Stdout)
	assert.NoError(t, err)
}

func TestLocalReportHTML(t *testing.T) {
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.RunDir = base.CompletePath("../", "testdata/assertinputs")
	rOpts.OutputFormat = HTMLOutputFormatKey

	err := rOpts.LocalRun(os.Stdout)
	assert.NoError(t, err)
}

func TestKubeReportText(t *testing.T) {
	base.SetupWithMock(t)
	// fix rOpts
	rOpts := NewReportOpts(driver.NewFakeKubeDriver(cli.New()))
	rOpts.Revision = 1

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", "experiment.yaml"))
	rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		StringData: map[string]string{"experiment.yaml": string(byteArray)},
	}, metav1.CreateOptions{})

	byteArray, _ = ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", "result.yaml"))
	rOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-result",
			Namespace: "default",
		},
		StringData: map[string]string{"result.yaml": string(byteArray)},
	}, metav1.CreateOptions{})

	err := rOpts.KubeRun(os.Stdout)
	assert.NoError(t, err)
}
