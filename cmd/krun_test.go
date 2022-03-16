package cmd

import (
	"context"
	"io/ioutil"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ia "github.com/iter8-tools/iter8/action"
	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

// Credit: this test structure is inspired by
// https://github.com/helm/helm/blob/main/cmd/helm/install_test.go
func TestKRun(t *testing.T) {
	ia.SetupWithMock(t)

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	kd.Revision = 1
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata/assertinputs", "experiment.yaml"))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"experiment.yaml": byteArray,
		},
	}, metav1.CreateOptions{})

	kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	tests := []cmdTestCase{
		// k report
		{
			name:   "k run",
			cmd:    "k run -g default --revision 1 --namespace default",
			golden: base.CompletePath("../testdata", "output/krun.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
