package base

import (
	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// initKubeFake initialize the Kube clientset with a fake
func initDynamicKubeFake(kd *DynamicKubeDriver, objects ...runtime.Object) {
	kd.dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
}

// initFake initializes fake Kubernetes and Helm clients
func initDynamicFake(kd *DynamicKubeDriver, objects ...runtime.Object) error {
	initDynamicKubeFake(kd, objects...)
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewDynamicFakeKubeDriver(s *cli.EnvSettings, objects ...runtime.Object) *DynamicKubeDriver {
	kd := &DynamicKubeDriver{
		EnvSettings: s,
	}
	initDynamicFake(kd, objects...)
	return kd
}
