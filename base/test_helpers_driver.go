package base

import (
	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// initKubeFake initialize the Kube clientset with a fake
func initKubeFake(driver *KubeDriver, objects ...runtime.Object) {
	driver.dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
}

// initFake initializes fake Kubernetes and Helm clients
func initFake(driver *KubeDriver, objects ...runtime.Object) error {
	initKubeFake(driver, objects...)
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *cli.EnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
	}
	initFake(kd, objects...)
	return kd
}
