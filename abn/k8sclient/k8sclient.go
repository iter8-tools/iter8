package k8sclient

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var Client = KubeClient{}

// KubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeClient struct {
	// typedClient enables interaction with a Kubernetes cluster using structured types
	typedClient kubernetes.Interface
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

// initKube initializes the Kubernetes clientset
func (kd *KubeClient) Initialize() (err error) {
	if kd.dynamicClient == nil {

		// get rest config
		// see https://pkg.go.dev/k8s.io/client-go/tools/clientcmd
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		restConfig, err := kubeConfig.ClientConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		kd.typedClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get dynamic client
		kd.dynamicClient, err = dynamic.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes dynamic client")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}

	return nil
}

func (c *KubeClient) Typed() kubernetes.Interface {
	return c.typedClient
}

func (c *KubeClient) Dynamic() dynamic.Interface {
	return c.dynamicClient
}
