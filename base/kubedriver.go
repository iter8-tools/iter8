package base

import (
	"errors"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeDriver embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeDriver struct {
	// EnvSettings provides generic Kubernetes options
	*EnvSettings
	// Clientset enables interaction with a Kubernetes cluster
	Clientset kubernetes.Interface
	// RestConfig is REST configuration of a Kubernetes cluster
	RestConfig *rest.Config
	// DynamicClient enables unstructured interaction with a Kubernetes cluster
	DynamicClient dynamic.Interface
	// Namespace
	Namespace *string
}

type GetObjectFuncType func(*KubeDriver, *corev1.ObjectReference) (*unstructured.Unstructured, error)

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings:   s,
		Clientset:     nil,
		RestConfig:    nil,
		DynamicClient: nil,
		Namespace:     nil,
	}
	return kd
}

// initKube initializes the Kubernetes clientset
func (kd *KubeDriver) initKube() (err error) {
	if kd.Clientset == nil {
		// get REST config
		kd.RestConfig, err = kd.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		kd.Clientset, err = kubernetes.NewForConfig(kd.RestConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		kd.DynamicClient, err = dynamic.NewForConfig(kd.RestConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes dynamic client")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}

		if kd.Namespace == nil {
			kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&clientcmd.ConfigOverrides{},
			)
			ns, _, err := kubeconfig.Namespace()
			if err != nil {
				ns = "default"
			}
			kd.Namespace = StringPointer(ns)
		}
	}

	return nil
}
