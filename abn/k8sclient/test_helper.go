package k8sclient

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// NewFakeKubeClient returns a fake Kubernetes client that is able to manage secrets
// Includes dynamic client with Deployments as listed objects
// Used by test cases in several packages to define (global) k8sclient.Client for testing
func NewFakeKubeClient(s *cli.EnvSettings, objects ...runtime.Object) *KubeClient {
	fakeClient := &KubeClient{
		EnvSettings: s,
		// default other fields
	}

	// secretDataReactor sets the secret.Data field based on the values from secret.StringData
	// Credit: this function is adapted from https://github.com/creydr/go-k8s-utils
	var secretDataReactor = func(action ktesting.Action) (bool, runtime.Object, error) {
		secret, _ := action.(ktesting.CreateAction).GetObject().(*corev1.Secret)

		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		for k, v := range secret.StringData {
			secret.Data[k] = []byte(v)
		}

		return false, nil, nil
	}

	fc := fake.NewSimpleClientset(objects...)
	fc.PrependReactor("create", "secrets", secretDataReactor)
	fc.PrependReactor("update", "secrets", secretDataReactor)
	fakeClient.typedClient = fc

	fakeClient.dynamicClient = NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "apps", Version: "v1", Resource: "deployments"}: "DeploymentList",
			{Group: "", Version: "v1", Resource: "services"}:        "ServiceList",
		},
		objects...)

	fakeClient.gvrMapper = func(o *unstructured.Unstructured) (*schema.GroupVersionResource, error) {
		m := map[schema.GroupVersionKind]schema.GroupVersionResource{
			{Group: "apps", Version: "v1", Kind: "Deployment"}: {Group: "apps", Version: "v1", Resource: "deployments"},
			{Version: "v1", Kind: "Service"}:                   {Group: "", Version: "v1", Resource: "services"},
		}

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

		gvr, ok := m[gvk]
		if !ok {
			return nil, fmt.Errorf("no mapping known for %v", gvk)
		}

		return &gvr, nil
	}

	return fakeClient
}
