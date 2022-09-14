package autox

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"

	"helm.sh/helm/v3/pkg/cli"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// k8sClient is a global variable. Before use, it must be assigned and initalized
var k8sClient *kubeClient = newKubeClient(cli.New())

// kubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type kubeClient struct {
	// EnvSettings provides generic Kubernetes options
	*cli.EnvSettings

	// clientset enables interaction with a Kubernetes cluster using structured types
	clientset kubernetes.Interface

	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

func newKubeClient(s *cli.EnvSettings) *kubeClient {
	return &kubeClient{
		EnvSettings: s,
		// default other fields
	}
}

// init initializes the Kubernetes clientset
func (c *kubeClient) init() (err error) {
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

// func (c *kubeClient) typed() kubernetes.Interface {
// 	return c.clientset
// }

func (c *kubeClient) dynamic() dynamic.Interface {
	return c.dynamicClient
}
