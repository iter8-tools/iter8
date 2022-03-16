package action

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

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
	// fix aOpts
	aOpts := NewAssertOpts(driver.NewFakeKubeDriver(cli.New()))
	aOpts.RunDir = base.CompletePath("../", "testdata/assertinputsfail")
	aOpts.Conditions = []string{Completed, NoFailure, SLOs}
	aOpts.Timeout = 5 * time.Second

	ok, err := aOpts.LocalRun()
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestKubeReportText(t *testing.T) {
	SetupWithMock(t)
	// fix aOpts
	aOpts := NewAssertOpts(driver.NewFakeKubeDriver(cli.New()))
	aOpts.Revision = 1
	aOpts.Conditions = []string{Completed, NoFailure, SLOs}

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", "experiment.yaml"))
	aOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"experiment.yaml": byteArray,
		},
	}, metav1.CreateOptions{})

	byteArray, _ = ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", "result.yaml"))
	aOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-result",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"result.yaml": byteArray,
		},
	}, metav1.CreateOptions{})

	ok, err := aOpts.KubeRun()
	assert.True(t, ok)
	assert.NoError(t, err)
}
