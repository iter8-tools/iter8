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

// TestNoObject tests that task fails if the object is not present
func TestNoObject(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask("non-existant-pod").WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

// TestWithoutCondition tests the task succeeds when there are no conditions on the object
// It should be successful
// Also validates parsing of timeout
// Also validates setting of default namespace
func TestWithoutConditions(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithTimeout("20s").Build()
	runTaskTest(t, rTask, true, ns, pod)
}

// TestWithCondition tests that the task succeeds when the condition is present and True
func TestWithCondition(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, true, ns, pod)
}

// TestWithFalseCondition tests that the task fails when the condition is present and not True
func TestWithFalseCondition(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "False").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("Ready").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

// TestConditionNotPresent tests that the task fails when the condition is not present (but others are)
func TestConditionNotPresent(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithCondition("NotPresent").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

// TestInvalidTimeout tests that the task fails when the specified timeout is invalid (not parseable)
func TestInvalidTimeout(t *testing.T) {
	ns, nm := "default", "test-pod"
	pod := newPod(ns, nm).WithCondition("Ready", "True").Build()
	rTask := NewReadinessTask(nm).WithResource("pods").WithNamespace(ns).WithTimeout("timeout").Build()
	runTaskTest(t, rTask, false, ns, pod)
}

// UTILITY METHODS for all tests

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
