package k8sclient

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/cli"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Client is a global variable. Before use, it must be assigned and initalized
var Client KubeClient

// KubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeClient struct {
	// EnvSettings provides generic Kubernets and Helm options; while Helm is not needed
	// for A/B(/n) functionality, we use here since we do so in other places and it
	// provides an easy way to get the Kubernetes configuration whether in cluster or not.
	*cli.EnvSettings
	// typedClient enables interaction with a Kubernetes cluster using structured types
	typedClient kubernetes.Interface
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

func NewKubeClient(s *cli.EnvSettings) *KubeClient {
	return &KubeClient{
		EnvSettings: s,
		// default other fields
	}
}

// initKube initializes the Kubernetes clientset
func (c *KubeClient) Initialize() (err error) {
	if c.dynamicClient == nil {
		// get rest config
		restConfig, err := c.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		c.typedClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get dynamic client
		c.dynamicClient, err = dynamic.NewForConfig(restConfig)
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
