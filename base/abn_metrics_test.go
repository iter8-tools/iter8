package base

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestABNMetricsTask(t *testing.T) {

	kd := NewFakeKubeDriver(cli.New())
	byteArray, _ := ioutil.ReadFile(CompletePath("../../testdata", "abninputs/readtest.yaml"))
	s, _ := kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app,
			Namespace: "default",
		},
		StringData: map[string]string{"versionData.yaml": string(byteArray)},
	}, metav1.CreateOptions{})
	s.ObjectMeta.Labels = map[string]string{"foo": "bar"}
	kd.Clientset.CoreV1().Secrets("default").Update(context.TODO(), s, metav1.UpdateOptions{})

	task := &collectABNMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectABNMetrics),
		},
		With: ABNMetricsInputs{
			Application: "app",
		},
	}

	exp := &Experiment{
		Spec:   []Task{task},
		Result: &ExperimentResult{},
		driver: kd,
	}

	exp.initResults(1)

	err := task.run(exp)
	assert.NoError(t, err)

	// any other assertions
}
