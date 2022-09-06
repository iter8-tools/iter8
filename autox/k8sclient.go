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

var (
	k8sclient = newKubeDriver(cli.New())
)

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

// newKubeDriver creates and returns a new KubeDriver
func newKubeDriver(s *cli.EnvSettings) *kubeClient {
	kd := &kubeClient{
		EnvSettings:   s,
		clientset:     nil,
		dynamicClient: nil,
	}
	return kd
}

// init initializes the Kubernetes clientset
func (kd *kubeClient) init() (err error) {
	if kd.dynamicClient == nil {
		// get REST config
		restConfig, err := kd.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		// clientSet will be replaced with a Helm client
		kd.clientset, err = kubernetes.NewForConfig(restConfig)
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
