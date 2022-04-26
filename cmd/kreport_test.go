package cmd

import (
	"context"
	"io/ioutil"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestKReport(t *testing.T) {
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
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", id.ExperimentSpecPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentSpecPath: string(byteArray)},
	}, metav1.CreateOptions{})

	byteArray, _ = ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", id.ExperimentResultPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-result",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentResultPath: string(byteArray)},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
