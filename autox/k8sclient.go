package autox

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"

	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// KubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeClient struct {
	// EnvSettings provides generic Kubernetes options
	*cli.EnvSettings

	// clientset enables interaction with a Kubernetes cluster using structured types
	clientset kubernetes.Interface

	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

// NewKubeClient creates an empty KubeClient
func NewKubeClient(s *cli.EnvSettings) *KubeClient {
	return &KubeClient{
		EnvSettings: s,
		// default other fields
	}
}

// init initializes the Kubernetes clientset
func (c *KubeClient) init() (err error) {
	if c.dynamicClient == nil {
		// get rest config
		restConfig, err := c.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}

		// get clientset
		c.clientset, err = kubernetes.NewForConfig(restConfig)
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

func (c *KubeClient) dynamic() dynamic.Interface {
	return c.dynamicClient
}
