package base

import (
	"helm.sh/helm/v3/pkg/cli"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// initKubeFake initialize the Kube clientset with a fake
func initKubeFake(kd *KubeDriver, objects ...runtime.Object) {
	kd.dynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *cli.EnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
	}
	initKubeFake(kd, objects...)
	return kd
}
