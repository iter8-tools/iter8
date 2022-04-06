package base

import (
	"errors"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
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
	// Mapping enables Object to Resource
	Mapping ObjectMapping
}

type GetObjectFuncType func(*KubeDriver, *corev1.ObjectReference) (*unstructured.Unstructured, error)

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings:   s,
		Clientset:     nil,
		RestConfig:    nil,
		DynamicClient: nil,
		Mapping:       nil,
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
		kd.Mapping = &KubernetesObjectMapping{}
	}

	return nil
}

type ObjectMapping interface {
	toGVK(*corev1.ObjectReference) schema.GroupVersionKind
	toGVR(*corev1.ObjectReference) (schema.GroupVersionResource, error)
}

type KubernetesObjectMapping struct{}

func (om *KubernetesObjectMapping) toGVK(objRef *corev1.ObjectReference) schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(objRef.APIVersion, objRef.Kind)
}

// TODO error handling
func (om *KubernetesObjectMapping) toGVR(objRef *corev1.ObjectReference) (schema.GroupVersionResource, error) {
	gvk := om.toGVK(objRef)
	dc, err := discovery.NewDiscoveryClientForConfig(kd.RestConfig)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil // This has the rigth resoruce field
}
