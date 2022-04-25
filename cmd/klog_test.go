package cmd

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestKLog(t *testing.T) {
	tests := []cmdTestCase{
		// k assert
		{
			name:   "k log",
			cmd:    "k log",
			golden: base.CompletePath("../testdata", "output/klog.txt"),
		},
	}

	// mock the environment
	// fake kube cluster
	*kd = *id.NewFakeKubeDriver(settings)
	kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-job-8218s",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "default-job",
			},
		},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
