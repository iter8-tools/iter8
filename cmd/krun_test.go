package cmd

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/time"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
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
	byteArray, _ := ioutil.ReadFile(base.CompletePath("../testdata", id.ExperimentSpecPath))
	kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-spec",
			Namespace: "default",
		},
		StringData: map[string]string{id.ExperimentSpecPath: string(byteArray)},
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
		StringData: map[string]string{id.ExperimentResultPath: string(resultBytes)},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
