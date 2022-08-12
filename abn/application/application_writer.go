package application

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type ApplicationReaderWriter struct {
	Client kubernetes.Interface
}
