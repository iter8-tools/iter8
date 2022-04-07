package base

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type podBuilder corev1.Pod

func newPod(ns string, nm string) *podBuilder {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nm,
			Namespace: ns,
		},
	}
	return (*podBuilder)(pod)
}

func (p *podBuilder) Build() *unstructured.Unstructured {
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured((*corev1.Pod)(p))
	if err != nil {
		return nil
	}
	return &unstructured.Unstructured{Object: o}
}

func (p *podBuilder) WithCondition(typ string, value string) *podBuilder {
	c := corev1.PodCondition{Type: (corev1.PodConditionType(typ))}
	switch strings.ToLower(value) {
	case "true":
		c.Status = corev1.ConditionTrue
	case "false":
		c.Status = corev1.ConditionFalse
	default:
		c.Status = corev1.ConditionUnknown
	}
	p.Status.Conditions = append(p.Status.Conditions, c)

	return p
}

type readinessTaskBuilder readinessTask

func NewReadinessTask(name string) *readinessTaskBuilder {
	rTask := &readinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: readinessInputs{
			Name: name,
		},
	}

	return (*readinessTaskBuilder)(rTask)
}

func (t *readinessTaskBuilder) WithResource(resource string) *readinessTaskBuilder {
	t.With.Resource = resource
	return t
}

func (t *readinessTaskBuilder) WithNamespace(ns string) *readinessTaskBuilder {
	t.With.Namespace = StringPointer(ns)
	return t
}

func (t *readinessTaskBuilder) WithTimeout(timeout string) *readinessTaskBuilder {
	t.With.Timeout = StringPointer(timeout)
	return t
}

func (t *readinessTaskBuilder) WithCondition(condition string) *readinessTaskBuilder {
	t.With.Condition = &condition
	return t
}

func (t *readinessTaskBuilder) Build() *readinessTask {
	return (*readinessTask)(t)
}

// also validates parsing of timeout
// also validates setting of default namespace
func TestWithoutConditions(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithTimeout("20s").Build()
	runTaskTest(t, rTask, true, ns, pod)
}

func TestWithCondition(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, true, ns, pod)
}

func TestWithFalseCondition(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "False").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

func TestConditionNotPresent(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("NotPresent").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

func TestNoObject(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask("non-existant-pod").WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

func TestInvalidTimeout(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithTimeout("timeout").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

// runTaskTest creates fake cluster with pod and runs rTask
func runTaskTest(t *testing.T, rTask *readinessTask, success bool, ns string, pod *unstructured.Unstructured) {
	*kd = *NewFakeKubeDriver(NewEnvSettings())
	rs := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	_, err := kd.DynamicClient.Resource(rs).Namespace(ns).Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err, "get failed")

	err = rTask.run(&Experiment{
		Tasks:  []Task{rTask},
		Result: &ExperimentResult{},
	})
	if success {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
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
