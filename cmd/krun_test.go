package cmd

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	id "github.com/iter8-tools/iter8/driver"

	"github.com/iter8-tools/iter8/base"
)

func TestKRun(t *testing.T) {
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
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", "experiment.yaml"))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{"experiment.yaml": string(byteArray)},
	}, metav1.CreateOptions{})

	resultBytes, _ := yaml.Marshal(base.ExperimentResult{
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
		Iter8Version:      base.MajorMinor,
	})
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-result",
			Namespace: "default",
		},
		StringData: map[string]string{"result.yaml": string(resultBytes)},
	}, metav1.CreateOptions{})

	kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
