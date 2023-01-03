package watcher

import (
	"testing"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetApplicationNameFromObjectName(t *testing.T) {
	for _, tt := range []struct {
		inputname          string
		expectedoutputname string
	}{
		{"hello", "hello"},
		{"hello-candidate-1", "hello"},
		{"hello-candidate-2", "hello"},
		{"hello-candidate-0", "hello-candidate-0"},
	} {
		t.Run(tt.inputname, func(t *testing.T) {
			n := getApplicationNameFromObjectName(tt.inputname)
			assert.Equal(t, tt.expectedoutputname, n)
		})
	}
}

func TestGetVersion(t *testing.T) {
	for _, tt := range []struct {
		testLabel  string
		hasLabel   bool
		labelValue string
	}{
		{"no label", false, ""},
		{"has label", true, "version"},
	} {
		t.Run(tt.testLabel, func(t *testing.T) {
			v, ok := getVersion(newUnstructuredDeployment("default", "name", tt.labelValue))
			if tt.hasLabel {
				assert.True(t, ok)
				assert.Equal(t, tt.labelValue, v)
			}
		})
	}

}

func newUnstructuredDeployment(namespace, name, version string) *unstructured.Unstructured {
	labels := map[string]string{
		versionLabel: version,
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	condition := appsv1.DeploymentCondition{Type: "Available", Status: "True"}
	deployment.Status.Conditions = append(deployment.Status.Conditions, condition)

	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployment)
	if err != nil {
		return nil
	}
	return &unstructured.Unstructured{Object: o}
}

func newUnstructuredService(namespace, name, version string) *unstructured.Unstructured {
	labels := map[string]interface{}{}
	if version != "" {
		labels[versionLabel] = version
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
				"labels":    labels,
			},
			"spec": name,
		},
	}
}
