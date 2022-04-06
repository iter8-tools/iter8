package base

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNoCondition(t *testing.T) {
	// fake kube cluster
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	rs := schema.GroupVersionResource{Group: "", Version: "", Resource: "pods"}

	ns, nm := "default", "test-pod"
	uPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "/",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      nm,
			},
		},
	}

	_, err := kd.DynamicClient.Resource(rs).Namespace(ns).Create(context.Background(), uPod, metav1.CreateOptions{})
	assert.NoError(t, err, "get failed")

	// create task
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Kind:      "pod",
			Name:      nm,
			Namespace: StringPointer(ns),
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
	rs := schema.GroupVersionResource{Group: "", Version: "", Resource: "pods"}

	ns, nm := "default", "test-pod"
	uPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "/",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      nm,
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	_, err := kd.DynamicClient.Resource(rs).Namespace(ns).Create(context.Background(), uPod, metav1.CreateOptions{})
	assert.NoError(t, err, "get failed")

	// create task
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Kind:      "pod",
			Name:      nm,
			Namespace: StringPointer(ns),
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

	rs := schema.GroupVersionResource{Group: "", Version: "", Resource: "pods"}

	ns, nm := "default", "test-pod"
	uPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "/",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      nm,
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	_, err := kd.DynamicClient.Resource(rs).Namespace(ns).Create(context.Background(), uPod, metav1.CreateOptions{})
	assert.NoError(t, err, "create failed")

	// create task
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Kind:      "pod",
			Name:      nm,
			Namespace: StringPointer(ns),
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

func TestNoObject(t *testing.T) {

	// fake kube cluster
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	rs := schema.GroupVersionResource{Group: "", Version: "", Resource: "pods"}

	ns, nm := "default", "test-pod"
	uPod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "/",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      nm,
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	_, err := kd.DynamicClient.Resource(rs).Namespace(ns).Create(context.Background(), uPod, metav1.CreateOptions{})
	assert.NoError(t, err, "get failed")

	// create task
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Kind:      "pod",
			Name:      "non-existant-pod",
			Namespace: StringPointer(ns),
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
	assert.Error(t, err)
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
	pods := []corev1.Pod{
		{    // no status
		}, { // no conditions
			Status: corev1.PodStatus{},
		}, { // empty list of conditions
			Status: corev1.PodStatus{ //
				Conditions: []corev1.PodCondition{},
			},
		}, { // not matched condition
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodConditionType("unmatched-condition"),
					Status: corev1.ConditionTrue,
				}},
			},
		}, { // no condition value
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Type: corev1.PodConditionType("no-status"),
				}},
			},
		}, { // no condition type
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					// Type: corev1.PodConditionType("no-type"),
					Status: corev1.ConditionTrue,
				}},
			},
		}, { // matched condition but wrong value
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodConditionType("matched-condition"),
					Status: corev1.ConditionFalse,
				}},
			},
		}, { // matched condition - success !
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodConditionType("matched-condition"),
					Status: corev1.ConditionTrue,
				}},
			},
		},
	}

	check(t, &pods[0], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[1], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[2], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[3], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[4], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[5], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[6], false, "matched-condition", string(corev1.ConditionTrue))
	check(t, &pods[7], true, "matched-condition", string(corev1.ConditionTrue))
}

func check(t *testing.T, kObj *corev1.Pod, expectSuccess bool, condition string, value string) {
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(kObj)
	unstructuredObj := unstructured.Unstructured{Object: o}
	assert.NoError(t, err)

	conditionStatus, err := getConditionStatus(&unstructuredObj, condition)
	if expectSuccess {
		assert.NoError(t, err)
		assert.Equal(t, value, *conditionStatus)
	}

}

func TestKube(t *testing.T) {
	assert.NoError(t, kd.initKube())
	settings = NewEnvSettings()
	assert.Equal(t, "default", settings.Namespace())
}
