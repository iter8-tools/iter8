package base

import (
	"errors"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base/log"

	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/client-go/dynamic"
)

var (
	kd = NewDynamicKubeDriver(cli.New())
)

// DynamicKubeDriver embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type DynamicKubeDriver struct {
	// EnvSettings provides generic Kubernetes options
	*cli.EnvSettings
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
}

// NewDynamicKubeDriver creates and returns a new KubeDriver
func NewDynamicKubeDriver(s *cli.EnvSettings) *DynamicKubeDriver {
	kd := &DynamicKubeDriver{
		EnvSettings:   s,
		dynamicClient: nil,
	}
	return kd
}

// initKube initializes the Kubernetes clientset
func (kd *DynamicKubeDriver) initKube() (err error) {
	if kd.dynamicClient == nil {
		// get REST config
		restConfig, err := kd.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		kd.dynamicClient, err = dynamic.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes dynamic client")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}

	return nil
}
