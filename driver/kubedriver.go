package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	// Import to initialize client auth plugins.

	// auth import enables automated authentication to various hosted clouds
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/retry"

	helmerrors "github.com/pkg/errors"
	helmdriver "helm.sh/helm/v3/pkg/storage/driver"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// secretTimeout is max time to wait for secret ops
	secretTimeout = 60 * time.Second
	// retryInterval is the duration between retries
	retryInterval = 1 * time.Second
	// ManifestFile is the name of the Kubernetes manifest file
	ManifestFile = "manifest.yaml"
)

// KubeDriver embeds Helm and Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs and Helm APIs
type KubeDriver struct {
	// EnvSettings provides generic Kubernetes and Helm options
	*cli.EnvSettings
	// Clientset enables interaction with a Kubernetes cluster
	Clientset kubernetes.Interface
	// Configuration enables Helm-based interaction with a Kubernetes cluster
	*action.Configuration
	// Test is the test name
	Test string
	// revision is the revision of the test
	revision int
}

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *cli.EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings:   s,
		Test:          DefaultTestName,
		Configuration: nil,
		Clientset:     nil,
	}
	return kd
}

// InitKube initializes the Kubernetes clientset
func (kd *KubeDriver) InitKube() error {
	if kd.Clientset == nil {
		// get REST config
		restConfig, err := kd.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		kd.Clientset, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}
	return nil
}

// initHelm initializes the Helm configuration
func (kd *KubeDriver) initHelm() error {
	if kd.Configuration == nil {
		// getting kube config
		kd.Configuration = new(action.Configuration)
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := kd.Configuration.Init(kd.EnvSettings.RESTClientGetter(), kd.EnvSettings.Namespace(), helmDriver, log.Logger.Debugf); err != nil {
			e := errors.New("unable to get Helm client config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		log.Logger.Info("inited Helm config")
	}
	return nil
}

// initRevision initializes the latest revision
func (kd *KubeDriver) initRevision() error {
	// update revision to latest, if none is specified
	if kd.revision <= 0 {
		if rel, err := kd.getLastRelease(); err == nil && rel != nil {
			kd.revision = rel.Version
		} else {
			return err
		}
	}
	return nil
}

// Init initializes the KubeDriver
func (kd *KubeDriver) Init() error {
	if err := kd.InitKube(); err != nil {
		return err
	}
	if err := kd.initHelm(); err != nil {
		return err
	}
	return kd.initRevision()
}

// getLastRelease fetches the last release of an Iter8 experiment
func (kd *KubeDriver) getLastRelease() (*release.Release, error) {
	log.Logger.Debugf("fetching latest revision for experiment group %v", kd.Test)
	// getting last revision
	rel, err := kd.Configuration.Releases.Last(kd.Test)
	if err != nil {
		if helmerrors.Is(err, helmdriver.ErrReleaseNotFound) {
			log.Logger.Debugf("experiment release not found")
			return nil, nil
		}
		e := fmt.Errorf("unable to get latest revision for experiment group %v", kd.Test)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	return rel, nil
}

// getExperimentSecretName yields the name of the experiment secret
func (kd *KubeDriver) getExperimentSecretName() string {
	return fmt.Sprintf("%v", kd.Test)
}

// getSecretWithRetry attempts to get a Kubernetes secret with retries
func (kd *KubeDriver) getSecretWithRetry(name string) (sec *corev1.Secret, err error) {
	err1 := retry.OnError(
		wait.Backoff{
			Steps:    int(secretTimeout / retryInterval),
			Cap:      secretTimeout,
			Duration: retryInterval,
			Factor:   1.0,
			Jitter:   0.1,
		},
		func(err2 error) bool { // retry on specific failures
			return kerrors.ReasonForError(err2) == metav1.StatusReasonForbidden
		},
		func() (err3 error) {
			secretsClient := kd.Clientset.CoreV1().Secrets(kd.Namespace())
			sec, err3 = secretsClient.Get(context.Background(), name, metav1.GetOptions{})
			return err3
		},
	)
	if err1 != nil {
		err = fmt.Errorf("unable to get secret %v", name)
		log.Logger.WithStackTrace(err1.Error()).Error(err)
		return nil, err
	}
	return sec, nil
}

// getExperimentSecret gets the Kubernetes experiment secret
func (kd *KubeDriver) getExperimentSecret() (s *corev1.Secret, err error) {
	return kd.getSecretWithRetry(kd.getExperimentSecretName())
}

// Read experiment from secret
func (kd *KubeDriver) Read() (*base.Experiment, error) {
	s, err := kd.getExperimentSecret()
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment")
		return nil, errors.New("unable to read experiment")
	}

	b, ok := s.Data[base.ExperimentFile]
	if !ok {
		err = fmt.Errorf("unable to extract experiment; spec secret has no %v field", base.ExperimentFile)
		log.Logger.Error(err)
		return nil, err
	}

	return ExperimentFromBytes(b)
}

// Write writes a Kubernetes experiment
func (kd *KubeDriver) Write(exp *base.Experiment) error {
	// write to metrics server
	// get URL of metrics server from environment variable
	metricsServerURL, ok := os.LookupEnv(base.MetricsServerURL)
	if !ok {
		errorMessage := "could not look up METRICS_SERVER_URL environment variable"
		log.Logger.Error(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	err := base.PutExperimentResultToMetricsService(metricsServerURL, exp.Metadata.Namespace, exp.Metadata.Name, exp.Result)
	if err != nil {
		errorMessage := "could not write experiment result to metrics service"
		log.Logger.Error(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	return nil
}

// GetRevision gets the experiment revision
func (kd *KubeDriver) GetRevision() int {
	return kd.revision
}
