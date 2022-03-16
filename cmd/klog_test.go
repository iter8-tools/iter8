package cmd

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

// Credit: this test structure is inspired by
// https://github.com/helm/helm/blob/main/cmd/helm/install_test.go
func TestKLog(t *testing.T) {

	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	kd.Revision = 1
	kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job-8218s",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "default-1-job",
			},
		},
	}, metav1.CreateOptions{})

	tests := []cmdTestCase{
		// k assert
		{
			name:   "k log",
			cmd:    "k log",
			golden: base.CompletePath("../testdata", "output/klog.txt"),
		},
	}

	runTestActionCmd(t, tests)
}
