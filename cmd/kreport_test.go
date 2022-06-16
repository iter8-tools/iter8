package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestKReport(t *testing.T) {
	os.Chdir(t.TempDir())
	tests := []cmdTestCase{
		// k report
		{
			name:   "k report",
			cmd:    "k report",
			golden: base.CompletePath("../testdata", "output/kreport.txt"),
		},
	}

	// mock the environment
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", id.ExperimentPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentPath: string(byteArray)},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
