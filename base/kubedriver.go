package base

import (
	"context"
	"errors"
	"fmt"
	"strings"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	Clientset     kubernetes.Interface
	RestConfig    *rest.Config
	GetObjectFunc GetObjectFuncType
}

type GetObjectFuncType func(*KubeDriver, *corev1.ObjectReference) (*unstructured.Unstructured, error)

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
		// Group:         DefaultExperimentGroup,
		Clientset:     nil,
		RestConfig:    nil,
		GetObjectFunc: GetRealObject,
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
	}

	return nil
}

func GetFakeObject(kd *KubeDriver, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	if strings.EqualFold("pod", objRef.Kind) {
		pod, err := kd.Clientset.CoreV1().Pods(objRef.Namespace).Get(context.Background(), objRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		if err != nil {
			return nil, err
		}
		return &unstructured.Unstructured{Object: obj}, nil

	}
	return nil, fmt.Errorf("resource %s not supported", objRef.Kind)
}

// getObject finds the object referenced by objRef using the client config restConfig
// uses the dynamic client; ie, retuns an unstructured object
// based on https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func GetRealObject(kd *KubeDriver, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	if objRef == nil {
		return nil, errors.New("nil object reference")
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(kd.RestConfig)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(kd.RestConfig)
	if err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(objRef.APIVersion, objRef.Kind)

	// 3. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	// 4. Obtain REST interface for the GVR
	namespace := objRef.Namespace // recall that we always set this
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	obj, err := dr.Get(context.Background(), objRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}
