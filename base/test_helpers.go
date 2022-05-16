package base

import (
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/jarcoal/httpmock"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"

	"helm.sh/helm/v3/pkg/chartutil"
	helmfake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	helmdriver "helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// SetupWithMock mocks an HTTP endpoint and registers and cleanup function
func SetupWithMock(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))
	t.Cleanup(httpmock.DeactivateAndReset)
}

// mockDriver is a mock driver that can be used to run experiments
type mockDriver struct {
	*Experiment

	metricsTemplate *template.Template
}

// ReadSpec enables spec to be read from the mock secret
func (m *mockDriver) ReadSpec() (ExperimentSpec, error) {
	return m.Experiment.Tasks, nil
}

// ReadMetricsSpec enables metrics spec to be read from the mock driver
func (m *mockDriver) ReadMetricsSpec(provider string) (*template.Template, error) {
	return m.metricsTemplate, nil
}

// ReadResult enables results to be read from the mock driver
func (m *mockDriver) ReadResult() (*ExperimentResult, error) {
	return m.Experiment.Result, nil
}

// WriteResult enables results to be written from the mock driver
func (m *mockDriver) WriteResult(r *ExperimentResult) error {
	m.Experiment.Result = r
	return nil
}

// initKubeFake initialize the Kube clientset with a fake
func initKubeFake(kd *KubeDriver, objects ...runtime.Object) {
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
}

// initHelmFake initializes the Helm config with a fake
// Credit: this function is adapted from helm
// https://github.com/helm/helm/blob/e9abdc5efe11cdc23576c20c97011d452201cd92/pkg/action/action_test.go#L37
func initHelmFake(kd *KubeDriver) {
	registryClient, err := registry.NewClient()
	if err != nil {
		log.Logger.Error(err)
		return
	}

	kd.Configuration = &action.Configuration{
		Releases:       storage.Init(helmdriver.NewMemory()),
		KubeClient:     &helmfake.FailingKubeClient{PrintingKubeClient: helmfake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities:   chartutil.DefaultCapabilities,
		RegistryClient: registryClient,
		Log:            log.Logger.Debugf,
	}
}

// initFake initializes fake Kubernetes and Helm clients
func initFake(kd *KubeDriver, objects ...runtime.Object) error {
	initKubeFake(kd, objects...)
	initHelmFake(kd)
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *cli.EnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
		Group:       DefaultExperimentGroup,
	}
	initFake(kd, objects...)
	return kd
}
