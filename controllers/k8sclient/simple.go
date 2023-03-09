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

type Client struct {
	*kubernetes.Clientset
	*dynamic.DynamicClient
}

// New creates a new kubernetes client
func New(settings *cli.EnvSettings) (*Client, error) {
	log.Logger.Trace("kubernetes client creation invoked ...")

	// get rest config
	restConfig, err := settings.RESTClientGetter().ToRESTConfig()
	if err != nil {
		e := errors.New("unable to get Kubernetes REST config")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	// get clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		e := errors.New("unable to get Kubernetes clientset")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	// get dynamic client
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		e := errors.New("unable to get Kubernetes dynamic client")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}

	log.Logger.Trace("returning kubernetes client ... ")

	return &Client{
		Clientset:     clientset,
		DynamicClient: dynamicClient,
	}, nil

}
