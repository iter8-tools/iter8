package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base"
	id "github.com/iter8-tools/iter8/driver"
	"helm.sh/helm/v3/pkg/time"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestKAssert(t *testing.T) {
	tests := []cmdTestCase{
		// k launch
		{
			name:   "k launch",
			cmd:    fmt.Sprintf("k launch -c load-test-http --chartsParentDir %v --noDownload --set url=https://httpbin.org/get --set duration=2s", base.CompletePath("../", "")),
			golden: base.CompletePath("../testdata", "output/klaunch.txt"),
		},
		// k run
		{
			name: "k run",
			cmd:  "k run -g default --namespace default",
		},
		// k assert
		{
			name:   "k assert",
			cmd:    "k assert -c completed -c nofailure -c slos",
			golden: base.CompletePath("../testdata", "output/kassert.txt"),
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

	kd.Clientset.BatchV1().Jobs("default").Create(context.TODO(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-1-job",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})

	runTestActionCmd(t, tests)
}
