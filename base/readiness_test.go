package base

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNoConditionObjectExists(t *testing.T) {

	// fake kube cluster
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	// kd.Revision = 1
	_, err := kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}, metav1.CreateOptions{})
	assert.NoError(t, err, "create failed")

	_, err = kd.Clientset.CoreV1().Pods("default").Get(context.Background(), "test-pod", metav1.GetOptions{})
	assert.NoError(t, err, "get failed")

	// create task
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Kind:      "pod",
			Name:      "test-pod",
			Namespace: StringPointer("default"),
		},
	}

	// create experiment
	exp := &Experiment{
		Tasks:  []Task{rTask},
		Result: &ExperimentResult{},
	}

	// run task
	err = rTask.run(exp)
	assert.NoError(t, err)
}
