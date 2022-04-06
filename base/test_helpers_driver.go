package base

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

// initKubeFake initialize the Kube clientset with a fake
func (kd *KubeDriver) initKubeFake(objects ...runtime.Object) {
	fc := fake.NewSimpleClientset(objects...)
	kd.Clientset = fc
	kd.DynamicClient = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	kd.Mapping = &FakeKubernetesObjectMapping{}
}

type FakeKubernetesObjectMapping struct{}

func (om *FakeKubernetesObjectMapping) toGVK(objRef *corev1.ObjectReference) schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(objRef.APIVersion, objRef.Kind)
}

func (om *FakeKubernetesObjectMapping) toGVR(objRef *corev1.ObjectReference) (schema.GroupVersionResource, error) {
	gvk := om.toGVK(objRef)
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: gvk.Kind + "s", // THIS IS WRONG  Kind is "pod", Resource is "pods"
	}, nil
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
