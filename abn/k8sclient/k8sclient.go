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

var (
	Client = NewKubeClient(cli.New())
)

// KubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeClient struct {
	// EnvSettings provides generic Kubernetes options
	*cli.EnvSettings
	// typedClient enables interaction with a Kubernetes cluster using structured types
	typedClient kubernetes.Interface
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

// NewKubeClient creates and returns a new KubeClient
func NewKubeClient(s *cli.EnvSettings) *KubeClient {
	return &KubeClient{
		EnvSettings:   s,
		typedClient:   nil,
		dynamicClient: nil,
	}
}

// initKube initializes the Kubernetes clientset
func (kd *KubeClient) Initialize() (err error) {
	if kd.dynamicClient == nil {
		// get REST config
		restConfig, err := kd.EnvSettings.RESTClientGetter().ToRESTConfig()
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
