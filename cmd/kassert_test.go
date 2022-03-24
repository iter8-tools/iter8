package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	id "github.com/iter8-tools/iter8/driver"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/iter8-tools/iter8/base"
)

func TestKAssert(t *testing.T) {
	srv := id.SetupWithRepo(t)
	base.SetupWithMock(t)
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", "experiment.yaml"))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-spec",
			Namespace: "default",
		},
		StringData: map[string]string{"experiment.yaml": string(byteArray)},
	}, metav1.CreateOptions{})

	kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	tests := []cmdTestCase{
		// k launch
		{
			name:   "init: k launch",
			cmd:    fmt.Sprintf("k launch -c load-test-http --repoURL %v --set url=https://httpbin.org/get --set duration=2s", srv.URL()),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// k run
		{
			name: "init: k run",
			cmd:  "k run -g default --revision 1 --namespace default",
		},
		// k assert
		{
			name:   "k assert",
			cmd:    "k assert -c completed -c nofailure -c slos",
			golden: base.CompletePath("../testdata", "output/kassert.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
