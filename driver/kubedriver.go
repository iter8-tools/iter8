package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Import to initialize client auth plugins.
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/downloader"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/retry"

	helmerrors "github.com/pkg/errors"
	helmdriver "helm.sh/helm/v3/pkg/storage/driver"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

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
func (driver *KubeDriver) initRevision() error {
	// update revision to latest, if none is specified
	if driver.revision <= 0 {
		if rel, err := driver.getLastRelease(); err == nil && rel != nil {
			driver.revision = rel.Version
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
	if err := kd.initRevision(); err != nil {
		return err
	}
	return nil
}

// getLastRelease fetches the last release of an Iter8 experiment
func (driver *KubeDriver) getLastRelease() (*release.Release, error) {
	log.Logger.Debugf("fetching latest revision for experiment group %v", driver.Group)
	// getting last revision
	rel, err := driver.Configuration.Releases.Last(driver.Group)
	if err != nil {
		if helmerrors.Is(err, helmdriver.ErrReleaseNotFound) {
			log.Logger.Debugf("experiment release not found")
			return nil, nil
		} else {
			e := fmt.Errorf("unable to get latest revision for experiment group %v", driver.Group)
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return nil, e
		}
	}
	return rel, nil
}

// getSpecSecretName yields the name of the experiment spec secret
func (driver *KubeDriver) getSpecSecretName() string {
	return fmt.Sprintf("%v-spec", driver.Group)
}

// getMetricsSpecSecretName yields the name of the experiment metrics spec
// secret
func (driver *KubeDriver) getMetricsSpecSecretName() string {
	return fmt.Sprintf("%v-metrics-spec", driver.Group)
}

// getResultSecretName yields the name of the experiment result secret
func (driver *KubeDriver) getResultSecretName() string {
	return fmt.Sprintf("%v-result", driver.Group)
}

// getExperimentJobName yields the name of the experiment job
func (driver *KubeDriver) getExperimentJobName() string {
	return fmt.Sprintf("%v-%v-job", driver.Group, driver.revision)
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

// getExperimentSpecSecret gets the Kubernetes experiment spec secret
func (driver *KubeDriver) getExperimentSpecSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getSpecSecretName())
}

// ReadSpec creates an ExperimentSpec struct for a Kubernetes experiment
func (driver *KubeDriver) ReadSpec() (base.ExperimentSpec, error) {
	s, err := driver.getExperimentSpecSecret()
	if err != nil {
		return nil, err
	}

	spec, ok := s.Data[ExperimentSpecPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment spec; spec secret has no %v field", ExperimentSpecPath)
		log.Logger.Error(err)
		return nil, err
	}

	return SpecFromBytes(spec)
}

// getMetricsSpecSecret gets the Kubernetes metrics spec secret
func (driver *KubeDriver) getMetricsSpecSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getMetricsSpecSecretName())
}

// ReadMetricsSpec creates a MetricsSpec struct for a Kubernetes experiment
func (driver *KubeDriver) ReadMetricsSpec(provider string) (*base.MetricsSpec, error) {
	s, err := driver.getMetricsSpecSecret()
	if err != nil {
		return nil, err
	}

	metricsSpecPath := provider + ExperimentMetricsPathSuffix
	res, ok := s.Data[metricsSpecPath]
	if !ok {
		err = fmt.Errorf("unable to extract metrics sepc; spec secret has no %v field", metricsSpecPath)
		log.Logger.Error(err)
		return nil, err
	}

	return MetricsSpecFromBytes(res)
}

// getExperimentResultSecret gets the Kubernetes experiment result secret
func (driver *KubeDriver) getExperimentResultSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getResultSecretName())
}

// ReadResult creates an ExperimentResult struct for a Kubernetes experiment
func (driver *KubeDriver) ReadResult() (*base.ExperimentResult, error) {
	s, err := driver.getExperimentResultSecret()
	if err != nil {
		return nil, err
	}

	res, ok := s.Data[ExperimentResultPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment result; result secret has no %v field", ExperimentResultPath)
		log.Logger.Error(err)
		return nil, err
	}

	return ResultFromBytes(res)
}

// PayloadValue is used to patch Kubernetes resources
type PayloadValue struct {
	// Op indicates the type of patch
	Op string `json:"op"`
	// Path is the JSON field path
	Path string `json:"path"`
	// Value is the value of the field
	Value string `json:"value"`
}

// formResultSecret creates the result secret using the result
func (driver *KubeDriver) formResultSecret(r *base.ExperimentResult) (*corev1.Secret, error) {
	byteArray, _ := yaml.Marshal(r)
	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: driver.getResultSecretName(),
		},
		StringData: map[string]string{ExperimentResultPath: string(byteArray)},
	}
	// formed result secret ...
	return &sec, nil
}

// updateExperimentResultSecret updates the experiment result secret
// as opposed to patch, update is an atomic operation
func (driver *KubeDriver) updateExperimentResultSecret(r *base.ExperimentResult) error {
	if sec, err := driver.formResultSecret(r); err == nil {
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

// WriteResult writes results for a Kubernetes experiment
func (driver *KubeDriver) WriteResult(r *base.ExperimentResult) error {
	if err := driver.updateExperimentResultSecret(r); err != nil {
		return err
	}
	return nil
}

// UpdateChartDependencies for an Iter8 experiment chart
// for now this function has one purpose ...
// bring iter8lib dependency into other experiment charts like load-test-http
func UpdateChartDependencies(chartDir string, settings *cli.EnvSettings) error {
	// client and settings may not really be initialized with proper values
	// should be ok considering iter8lib is a local file dependency
	if settings == nil {
		settings = cli.New()
	}
	client := action.NewDependency()
	man := &downloader.Manager{
		Out:              ioutil.Discard,
		ChartPath:        chartDir,
		Keyring:          client.Keyring,
		SkipUpdate:       client.SkipRefresh,
		Getters:          getter.All(settings),
		RepositoryConfig: settings.RepositoryConfig,
		RepositoryCache:  settings.RepositoryCache,
		Debug:            settings.Debug,
	}
	log.Logger.Debug("updating chart ", chartDir)
	if err := man.Update(); err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to update chart dependencies")
		return err
	}
	return nil
}

// writeManifest writes the Kubernetes experiment manifest to a local file
func writeManifest(rel *release.Release) error {
	err := ioutil.WriteFile(ManifestFile, []byte(rel.Manifest), 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write kubernetes manifest into ", ManifestFile)
		return err
	}
	log.Logger.Info("wrote kubernetes manifest into ", ManifestFile)
	return nil
}

// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/upgrade.go#L69
// Upgrade a Kubernetes experiment to the next release
func (driver *KubeDriver) upgrade(chartDir string, valueOpts values.Options, group string, dry bool) error {
	client := action.NewUpgrade(driver.Configuration)
	client.Namespace = driver.Namespace()
	client.DryRun = dry

	ch, vals, err := driver.getChartAndVals(chartDir, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and vals for %v", chartDir)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// Create context and prepare the handle of SIGTERM
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		fmt.Printf("experiment for group %s has been cancelled.\n", group)
		cancel()
	}()

	rel, err := client.RunWithContext(ctx, group, ch, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// upgrading revision info
	driver.revision = rel.Version

	// write manifest if dry
	if dry {
		err := writeManifest(rel)
		if err != nil {
			return err
		}
		log.Logger.Info("dry run complete")
	} else {
		log.Logger.Info("experiment launched. Happy Iter8ing!")
	}

	return nil
}

// install a Kubernetes experiment
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (driver *KubeDriver) install(chartDir string, valueOpts values.Options, group string, dry bool) error {
	client := action.NewInstall(driver.Configuration)
	client.Namespace = driver.Namespace()
	client.DryRun = dry
	client.ReleaseName = group

	ch, vals, err := driver.getChartAndVals(chartDir, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and vals for %v", chartDir)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// Create context and prepare the handle of SIGTERM
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		fmt.Printf("experiment for group %s has been cancelled.\n", group)
		cancel()
	}()

	rel, err := client.RunWithContext(ctx, ch, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// upgrading revision info
	driver.revision = rel.Version

	// write manifest if dry
	if dry {
		err := writeManifest(rel)
		if err != nil {
			return err
		}
		log.Logger.Info("dry run complete")
	} else {
		log.Logger.Info("experiment launched. Happy Iter8ing!")
	}

	return nil
}

// Launch a Kubernetes experiment
func (driver *KubeDriver) Launch(chartDir string, valueOpts values.Options, group string, dry bool) error {
	if driver.revision <= 0 {
		return driver.install(chartDir, valueOpts, group, dry)
	} else {
		return driver.upgrade(chartDir, valueOpts, group, dry)
	}
}

// Delete a Kubernetes experiment group
func (driver *KubeDriver) Delete() error {
	client := action.NewUninstall(driver.Configuration)
	_, err := client.Run(driver.Group)
	if err != nil {
		e := fmt.Errorf("deletion of experiment group %v failed", driver.Group)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	log.Logger.Infof("experiment group %v deleted", driver.Group)
	return nil
}

// getChartAndVals gets experiment chart and its values
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (driver *KubeDriver) getChartAndVals(chartDir string, valueOpts values.Options) (*chart.Chart, map[string]interface{}, error) {
	// update dependencies for the chart
	if err := UpdateChartDependencies(chartDir, driver.EnvSettings); err != nil {
		return nil, nil, err
	}

	// form chart values
	p := getter.All(driver.EnvSettings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		e := fmt.Errorf("unable to merge chart values")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, nil, e
	}

	// attempt to load the chart
	ch, err := loader.Load(chartDir)
	if err != nil {
		e := fmt.Errorf("unable to load chart")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, nil, e
	}

	if err := checkIfInstallable(ch); err != nil {
		return nil, nil, err
	}

	if ch.Metadata.Deprecated {
		log.Logger.Warning("this chart is deprecated")
	}
	return ch, vals, nil
}

// checkIfInstallable validates if a chart can be installed
// Only application chart type is installable
// Credit: this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L270
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	e := fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
	log.Logger.Error(e)
	return e
}

// GetExperimentLogs gets logs for a Kubernetes experiment
func (driver *KubeDriver) GetExperimentLogs() (string, error) {
	podsClient := driver.Clientset.CoreV1().Pods(driver.Namespace())
	pods, err := podsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%v", driver.getExperimentJobName()),
	})
	if err != nil {
		e := errors.New("unable to get experiment pod(s)")
		log.Logger.Error(e)
		return "", e
	}
	lgs := make([]string, len(pods.Items))
	for i, p := range pods.Items {
		req := podsClient.GetLogs(p.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			e := errors.New("error in opening log stream")
			log.Logger.Error(e)
			return "", e
		}

		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			e := errors.New("error in copy information from podLogs to buf")
			log.Logger.Error(e)
			return "", e
		}
		str := buf.String()
		lgs[i] = str
	}
	return strings.Join(lgs, "\n***\n"), nil
}
