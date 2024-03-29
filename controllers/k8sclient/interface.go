// Package k8sclient provides the Kubernetes client for the controllers package
package k8sclient

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Interface enables interaction with a Kubernetes cluster
// Can be mocked in unit tests with fake implementation
type Interface interface {
	kubernetes.Interface
	dynamic.Interface
	Patch(gvr schema.GroupVersionResource, objNamespace string, objName string, by []byte) (*unstructured.Unstructured, error)
}
