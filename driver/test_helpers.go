package driver

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	helmfake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage"
	helmdriver "helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// initKube initialize the Kube clientset with a fake
func (kd *KubeDriver) initKubeFake(objects ...runtime.Object) {
	// secretDataReactor sets the secret.Data field based on the values from secret.StringData
	// Credit: this function is adapted from https://github.com/creydr/go-k8s-utils
	var secretDataReactor = func(action ktesting.Action) (bool, runtime.Object, error) {
		secret, _ := action.(ktesting.CreateAction).GetObject().(*corev1.Secret)

		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		for k, v := range secret.StringData {
			sEnc := base64.StdEncoding.EncodeToString([]byte(v))
			secret.Data[k] = []byte(sEnc)
		}

		return false, nil, nil
	}

	fc := fake.NewSimpleClientset(objects...)
	fc.PrependReactor("create", "secrets", secretDataReactor)
	kd.Clientset = fc
}

// initHelmFake initializes the Helm config with a fake
// Credit: this function is adapted from helm
// https://github.com/helm/helm/blob/e9abdc5efe11cdc23576c20c97011d452201cd92/pkg/action/action_test.go#L37
func (kd *KubeDriver) initHelmFake() {
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
func (driver *KubeDriver) initFake(objects ...runtime.Object) error {
	driver.initKubeFake(objects...)
	driver.initHelmFake()
	return nil
}

// NewFakeKubeDriver creates and returns a new KubeDriver with fake clients
func NewFakeKubeDriver(s *cli.EnvSettings, objects ...runtime.Object) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings: s,
		Group:       DefaultExperimentGroup,
	}
	kd.initFake(objects...)
	return kd
}
