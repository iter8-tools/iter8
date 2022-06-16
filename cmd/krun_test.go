package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKRun(t *testing.T) {
	os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// k report
		{
			name:   "k run",
			cmd:    "k run -g default --namespace default",
			golden: base.CompletePath("../testdata", "output/krun.txt"),
		},
	}

	// mock the environment
	base.SetupWithMock(t)
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", id.ExperimentPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
