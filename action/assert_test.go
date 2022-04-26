package action

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLocalAssert(t *testing.T) {
	// fix aOpts
	aOpts := NewAssertOpts(driver.NewFakeKubeDriver(cli.New()))
	aOpts.RunDir = base.CompletePath("../", "testdata/assertinputs")
	aOpts.Conditions = []string{Completed, NoFailure, SLOs}

	ok, err := aOpts.LocalRun()
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestLocalAssertFailing(t *testing.T) {
	// fix aOpts
	aOpts := NewAssertOpts(driver.NewFakeKubeDriver(cli.New()))
	aOpts.RunDir = base.CompletePath("../", "testdata/assertinputsfail")
	aOpts.Conditions = []string{Completed, NoFailure, SLOs}
	aOpts.Timeout = 5 * time.Second

	ok, err := aOpts.LocalRun()
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestKubeAssert(t *testing.T) {
	// fix aOpts
	aOpts := NewAssertOpts(driver.NewFakeKubeDriver(cli.New()))
	aOpts.Conditions = []string{Completed, NoFailure, SLOs}

	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", driver.ExperimentSpecPath))
	aOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentSpecPath: string(byteArray)},
	}, metav1.CreateOptions{})

	byteArray, _ = ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", driver.ExperimentResultPath))
	aOpts.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-result",
			Namespace: "default",
		},
		StringData: map[string]string{driver.ExperimentResultPath: string(byteArray)},
	}, metav1.CreateOptions{})

	ok, err := aOpts.KubeRun()
	assert.True(t, ok)
	assert.NoError(t, err)
}
