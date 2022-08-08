package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iter8-tools/iter8/base/log"

	"helm.sh/helm/v3/pkg/cli"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

const (
	// secretTimeout is max time to wait for secret ops
	secretTimeout = 60 * time.Second
	// retryInterval is the duration between retries
	retryInterval = 1 * time.Second
	// ManifestFile is the name of the Kubernetes manifest file
	ManifestFile = "manifest.yaml"
)

// copy of common.go
const (
	// ExperimentPath is the name of the experiment file
	ExperimentPath = "experiment.yaml"
	// DefaultExperimentGroup is the name of the default experiment chart
	DefaultExperimentGroup = "default"
)

// ExperimentFromBytes reads experiment from bytes
func ExperimentFromBytes(b []byte) (*Experiment, error) {
	e := Experiment{}
	err := yaml.Unmarshal(b, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment")
		return nil, err
	}
	return &e, err
}

// end copy of common.go

var (
	kd = NewKubeDriver(cli.New())
)

// KubeDriver embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeDriver struct {
	// EnvSettings provides generic Kubernetes options
	*cli.EnvSettings
	// Clientset enables interaction with a Kubernetes cluster using structured types
	Clientset kubernetes.Interface
	// dynamicClient enables unstructured interaction with a Kubernetes cluster
	DynamicClient dynamic.Interface
	// Group is the experiment group
	Group string
	// revision is the revision of the experiment
	revision int
}

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *cli.EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings:   s,
		Group:         DefaultExperimentGroup,
		Clientset:     nil,
		DynamicClient: nil,
	}
	return kd
}

// initKube initializes the Kubernetes clientset
func (kd *KubeDriver) InitKube() (err error) {
	if kd.DynamicClient == nil {
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
		// get dynamic client
		kd.DynamicClient, err = dynamic.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes dynamic client")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}

	return nil
}

// getExperimentSecretName yields the name of the experiment secret
func (driver *KubeDriver) getExperimentSecretName() string {
	return fmt.Sprintf("%v", driver.Group)
}

// getSecretWithRetry attempts to get a Kubernetes secret with retries
func (driver *KubeDriver) getSecretWithRetry(name string) (sec *corev1.Secret, err error) {
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
			secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
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
func (driver *KubeDriver) getExperimentSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getExperimentSecretName())
}

// Read experiment from secret
func (driver *KubeDriver) Read() (*Experiment, error) {
	s, err := driver.getExperimentSecret()
	if err != nil {
		return nil, err
	}

	b, ok := s.Data[ExperimentPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment; spec secret has no %v field", ExperimentPath)
		log.Logger.Error(err)
		return nil, err
	}

	return ExperimentFromBytes(b)
}

// formExperimentSecret creates the experiment secret using the experiment
func (driver *KubeDriver) formExperimentSecret(e *Experiment) (*corev1.Secret, error) {
	byteArray, _ := yaml.Marshal(e)
	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: driver.getExperimentSecretName(),
			Annotations: map[string]string{
				"iter8.tools/group": driver.Group,
			},
		},
		StringData: map[string]string{ExperimentPath: string(byteArray)},
	}
	// formed experiment secret ...
	return &sec, nil
}

// updateExperimentSecret updates the experiment secret
// as opposed to patch, update is an atomic operation
func (driver *KubeDriver) updateExperimentSecret(e *Experiment) error {
	if sec, err := driver.formExperimentSecret(e); err == nil {
		secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
		_, err1 := secretsClient.Update(context.Background(), sec, metav1.UpdateOptions{})
		// TODO: Evaluate if result secret update requires retries.
		// Probably not. Conflicts will be avoided if cronjob avoids parallel jobs.
		if err1 != nil {
			err2 := fmt.Errorf("unable to update secret %v", sec.Name)
			log.Logger.WithStackTrace(err1.Error()).Error(err2)
			return err2
		}
	} else {
		return err
	}
	return nil
}

// Write writes a Kubernetes experiment
func (driver *KubeDriver) Write(e *Experiment) error {
	if err := driver.updateExperimentSecret(e); err != nil {
		return err
	}
	return nil
}

// GetRevision gets the experiment revision
func (driver *KubeDriver) GetRevision() int {
	return driver.revision
}
