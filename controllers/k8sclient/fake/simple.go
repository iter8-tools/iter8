package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

type Client struct {
	*fakek8s.Clientset
	*fakedynamic.FakeDynamicClient
}

func New(objects ...runtime.Object) *Client {
	return &Client{
		fakek8s.NewSimpleClientset(objects...),
		fakedynamic.NewSimpleDynamicClient(runtime.NewScheme(), objects...),
	}
}
