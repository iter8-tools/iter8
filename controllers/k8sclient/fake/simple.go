// Package fake provides fake Kuberntes clients for testing
package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

// Client provides structured and dynamic fake clients
type Client struct {
	*fakek8s.Clientset
	*fakedynamic.FakeDynamicClient
}

// New returns a new fake Kubernetes client populated with runtime objects
func New(objects ...runtime.Object) *Client {
	return &Client{
		fakek8s.NewSimpleClientset(objects...),
		fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), objects...),
	}
}
