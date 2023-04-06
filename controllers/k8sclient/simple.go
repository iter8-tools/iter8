package k8sclient

import (
	"context"
	"errors"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/cli"

	// Import to initialize client auth plugins.
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Client provides typed and dynamic Kubernetes clients
type Client struct {
	// typed Kubernetes client
	*kubernetes.Clientset
	// dynamic Kubernetes client
	*dynamic.DynamicClient
}

func (cl *Client) Patch(gvr schema.GroupVersionResource, objNamespace string, objName string, jsonBytes []byte) (*unstructured.Unstructured, error) {
	return cl.DynamicClient.Resource(gvr).Namespace(objNamespace).Patch(context.TODO(), objName, types.ApplyPatchType, jsonBytes, metav1.PatchOptions{
		FieldManager: "iter8-controller",
		Force:        base.BoolPointer(true),
	})
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
