package base

import (

	// "helm.sh/helm/v3/pkg/action"
	// "helm.sh/helm/v3/pkg/chartutil"
	// "helm.sh/helm/v3/pkg/cli"
	// helmfake "helm.sh/helm/v3/pkg/kube/fake"
	// "helm.sh/helm/v3/pkg/registry"
	// "helm.sh/helm/v3/pkg/repo/repotest"
	// "helm.sh/helm/v3/pkg/storage"
	// helmdriver "helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"

	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	kd.DynamicClient = NewForClient(kd.Clientset)

	kd.Mapper = nil
	// get REST config
	restConfig, err := kd.MyEnvSettings.RESTClientGetter().ToRESTConfig()
	if err != nil {
		return
	}
	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return
	}
	kd.Mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

}

// // initHelmFake initializes the Helm config with a fake
// // Credit: this function is adapted from helm
// // https://github.com/helm/helm/blob/e9abdc5efe11cdc23576c20c97011d452201cd92/pkg/action/action_test.go#L37
// func (kd *KubeDriver) initHelmFake() {
// 	registryClient, err := registry.NewClient()
// 	if err != nil {
// 		log.Logger.Error(err)
// 		return
// 	}

// 	kd.Configuration = &action.Configuration{
// 		Releases:       storage.Init(helmdriver.NewMemory()),
// 		KubeClient:     &helmfake.FailingKubeClient{PrintingKubeClient: helmfake.PrintingKubeClient{Out: ioutil.Discard}},
// 		Capabilities:   chartutil.DefaultCapabilities,
// 		RegistryClient: registryClient,
// 		Log:            log.Logger.Debugf,
// 	}
// }

// initFake initializes fake Kubernetes and Helm clients
func (driver *KubeDriver) initFake(objects ...runtime.Object) error {
	driver.initKubeFake(objects...)
	// driver.initHelmFake()
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *MyEnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		MyEnvSettings: s,
		Group:         DefaultExperimentGroup,
	}
	kd.initFake(objects...)
	return kd
}

// // SetupWithRepo creates a local experiment chart repo and cleans up after test
// func SetupWithRepo(t *testing.T) *repotest.Server {
// 	srv, err := repotest.NewTempServerWithCleanup(t, CompletePath("../", "testdata/charts/*.tgz*"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Cleanup(srv.Stop)

// 	if err := srv.LinkIndices(); err != nil {
// 		t.Fatal(err)
// 	}
// 	return srv
// }

type MyEnvSettings struct {
	namespace string
	config    *genericclioptions.ConfigFlags
}

func NewEnvSettings() *MyEnvSettings {
	env := &MyEnvSettings{
		namespace: "",
		config:    &genericclioptions.ConfigFlags{},
	}

	return env
}

// Namespace gets the namespace from the configuration
func (s *MyEnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "default"
}

// SetNamespace sets the namespace in the configuration
func (s *MyEnvSettings) SetNamespace(namespace string) {
	s.namespace = namespace
}

// RESTClientGetter gets the kubeconfig from EnvSettings
func (s *MyEnvSettings) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.config
}
