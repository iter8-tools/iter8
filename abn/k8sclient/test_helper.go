package k8sclient

import (
	"helm.sh/helm/v3/pkg/cli"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// initKubeFake initialize the Kube clientset with a fake
func initKubeFake(kd *KubeClient, objects ...runtime.Object) {
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
	kd.typedClient = fc

	kd.dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
}

// initFake initializes fake Kubernetes and Helm clients
func initFake(kd *KubeClient, objects ...runtime.Object) error {
	initKubeFake(kd, objects...)
	return nil
}

// NewFakeKubeClient creates and returns a new KubeClient with fake clients
func NewFakeKubeClient(s *cli.EnvSettings, objects ...runtime.Object) *KubeClient {
	kd := &KubeClient{
		EnvSettings: s,
	}
	initFake(kd, objects...)
	return kd
}