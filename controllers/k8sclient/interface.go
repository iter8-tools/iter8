package k8sclient

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Interface enables interaction with a Kubernetes cluster
type Interface interface {
	kubernetes.Interface
	dynamic.Interface
}
