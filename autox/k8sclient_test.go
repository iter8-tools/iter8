package autox

import (
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// newFakeKubeClient returns a fake Kubernetes client that is able to manage secrets
// Includes dynamic client with Deployments as listed objects
// Used by test cases in several packages to define (global) k8sclient.Client for testing
func newFakeKubeClient(s *cli.EnvSettings, objects ...runtime.Object) *KubeClient {
	fakeClient := KubeClient{
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
	fakeClient.clientset = fc

	// fakeClient.dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	fakeClient.dynamicClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "apps", Version: "v1", Resource: "deployments"}: "DeploymentList",
		},
		objects...)

	return &fakeClient
}
