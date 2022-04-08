package base

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// initKubeFake initialize the Kube clientset with a fake
func (kd *KubeDriver) initKubeFake(objects ...runtime.Object) {
	kd.DynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	kd.Namespace = StringPointer("default")
}

// initFake initializes fake Kubernetes and Helm clients
func (driver *KubeDriver) initFake(objects ...runtime.Object) error {
	driver.initKubeFake(objects...)
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *EnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
	}
	kd.initFake(objects...)
	return kd
}

type EnvSettings struct {
	config *genericclioptions.ConfigFlags
}

func NewEnvSettings() *EnvSettings {
	env := &EnvSettings{
		config: &genericclioptions.ConfigFlags{},
	}

	return env
}

// Namespace gets the namespace from the configuration
func (s *EnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}

// RESTClientGetter gets the kubeconfig from EnvSettings
func (s *EnvSettings) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.config
}
