package base

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNoCondition(t *testing.T) {

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

func TestConditionPresent(t *testing.T) {

	// fake kube cluster
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	// kd.Revision = 1
	_, err := kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodConditionType("Ready"),
				Status: corev1.ConditionTrue,
			}},
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
			Condition: StringPointer("Ready"),
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

func TestConditionNotPresent(t *testing.T) {

	// fake kube cluster
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	// kd.Revision = 1
	_, err := kd.Clientset.CoreV1().Pods("default").Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodConditionType("Ready"),
				Status: corev1.ConditionTrue,
			}},
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
			Condition: StringPointer("NotPresent"),
		},
	}

	// create experiment
	exp := &Experiment{
		Tasks:  []Task{rTask},
		Result: &ExperimentResult{},
	}

	// run task
	err = rTask.run(exp)
	assert.Error(t, err)
	assert.Equal(t, "expected condition not found", err.Error())
}

func TestValidation(t *testing.T) {
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

	// invalid timeout
	rTask.With.Timeout = StringPointer("invalid")
	assert.Error(t, rTask.validateInputs())

	// valid timeout
	rTask.With.Timeout = StringPointer("3m5s")
	assert.NoError(t, rTask.validateInputs())
}

func TestGetConditionStatus(t *testing.T) {
	pod := corev1.Pod{
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodConditionType("false"),
				Status: corev1.ConditionFalse,
			}, {
				Type:   corev1.PodConditionType("true"),
				Status: corev1.ConditionTrue,
			}, {
				Type:   corev1.PodConditionType("unknown"),
				Status: corev1.ConditionUnknown,
			}},
		},
	}

	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	obj := unstructured.Unstructured{Object: o}
	assert.NoError(t, err)

	status, err := getConditionStatus(&obj, "true")
	assert.NoError(t, err)
	assert.Equal(t, "True", *status)

	status, err = getConditionStatus(&obj, "false")
	assert.NoError(t, err)
	assert.Equal(t, "False", *status)

	_, err = getConditionStatus(&obj, "invalid")
	assert.Error(t, err)
	assert.Equal(t, "expected condition not found", err.Error())
}
