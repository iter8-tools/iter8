package base

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// initKubeFake initialize the Kube clientset with a fake
func (kd *KubeDriver) initKubeFake(objects ...runtime.Object) {
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
	kd.Clientset = fc
	kd.GetObjectFunc = GetFakeObject
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
	namespace string
	config    *genericclioptions.ConfigFlags
}

func NewEnvSettings() *EnvSettings {
	env := &EnvSettings{
		namespace: "",
		config:    &genericclioptions.ConfigFlags{},
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
