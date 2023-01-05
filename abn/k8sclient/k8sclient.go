package k8sclient

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/cli"

	// Import to initialize client auth plugins.
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	// packages for cloud authentication
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Client is a global variable. Before use, it must be assigned and initialized
var Client = *NewKubeClient(cli.New())

// KubeClient embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeClient struct {
	// EnvSettings provides generic Kubernets and Helm options; while Helm is not needed
	// for A/B/n functionality, we use here since we do so in other places and it
	// provides an easy way to get the Kubernetes configuration whether in cluster or not.
	*cli.EnvSettings
	// typedClient enables interaction with a Kubernetes cluster using structured types
	typedClient kubernetes.Interface
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	dynamicClient dynamic.Interface
	gvrMapper     func(o *unstructured.Unstructured) (*schema.GroupVersionResource, error)
}

// NewKubeClient returns a KubeClient with the given settings.
// Must be initialized before it can be used.
func NewKubeClient(s *cli.EnvSettings) *KubeClient {
	return &KubeClient{
		EnvSettings: s,
		// default other fields
	}
}

// Initialize initializes the Kubernetes clientset
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
		c.gvrMapper = func(o *unstructured.Unstructured) (*schema.GroupVersionResource, error) {
			if o == nil {
				return nil, errors.New("no object provided")
			}

			// get GVK
			gv, err := schema.ParseGroupVersion(o.GetAPIVersion())
			if err != nil {
				return nil, err
			}
			gvk := schema.GroupVersionKind{
				Group:   gv.Group,
				Version: gv.Version,
				Kind:    o.GetKind(),
			}

			// convert GVK to GVR
			// see https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
			dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
			if err != nil {
				return nil, err
			}
			mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
			mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
			if err != nil {
				return nil, err
			}
			gvr := mapping.Resource

			return &gvr, nil
		}
	}

	return nil
}

// Typed is the typed k8s client interface
func (c *KubeClient) Typed() kubernetes.Interface {
	return c.typedClient
}

// Dynamic is the dynamic (untyped) k8s client interface
func (c *KubeClient) Dynamic() dynamic.Interface {
	return c.dynamicClient
}

// GVR determines the GroupVersionResource of an object
func (c *KubeClient) GVR(o *unstructured.Unstructured) (*schema.GroupVersionResource, error) {
	return c.gvrMapper(o)
}
