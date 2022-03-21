package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// maxGetRetries is the number of tries to retry while fetching Kubernetes objects
	maxGetRetries = 2
	// getRetryInterval is the duration between retrials
	getRetryInterval = 1 * time.Second
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
	// Revision is the revision of the experiment
	Revision int
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

// initKube initializes the Kubernetes clientset
func (kd *KubeDriver) initKube() error {
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
	if driver.Revision <= 0 {
		if rel, err := driver.getLastRelease(); err == nil && rel != nil {
			driver.Revision = rel.Version
		} else {
			return err
		}
	}
	return nil
}

// Init initializes the KubeDriver
func (driver *KubeDriver) Init() error {
	if err := driver.initKube(); err != nil {
		return err
	}
	if err := driver.initHelm(); err != nil {
		return err
	}
	if err := driver.initRevision(); err != nil {
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
	return fmt.Sprintf("%v-%v-spec", driver.Group, driver.Revision)
}

// getResultSecretName yields the name of the experiment result secret
func (driver *KubeDriver) getResultSecretName() string {
	return fmt.Sprintf("%v-%v-result", driver.Group, driver.Revision)
}

// getExperimentJobName yields the name of the experiment job
func (driver *KubeDriver) getExperimentJobName() string {
	return fmt.Sprintf("%v-%v-job", driver.Group, driver.Revision)
}

// getSecretWithRetry attempts to get a Kubernetes secret with retry
func (driver *KubeDriver) getSecretWithRetry(name string) (s *corev1.Secret, err error) {
	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	for i := 0; i < maxGetRetries; i++ {
		s, err = secretsClient.Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return s, err
		}
		if !k8serrors.IsNotFound(err) {
			log.Logger.Warningf("unable to get secret: %s; %s\n", name, err.Error())
		}
		time.Sleep(getRetryInterval)
	}
	e := fmt.Errorf("unable to get secret %v", name)
	log.Logger.Warning(e)
	return nil, e
}

// getJobWithRetry attempts to get a Kubernetes job with retry
func (driver *KubeDriver) getJobWithRetry(name string) (*batchv1.Job, error) {

	jobsClient := driver.Clientset.BatchV1().Jobs(driver.Namespace())

	for i := 0; i < maxGetRetries; i++ {
		j, err := jobsClient.Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return j, err
		}
		if !k8serrors.IsNotFound(err) {
			log.Logger.Warningf("unable to get job: %s; %s\n", name, err.Error())
		}
		time.Sleep(getRetryInterval)
	}
	e := fmt.Errorf("unable to get job %v", name)
	log.Logger.Error(e)
	return nil, e
}

// getExperimentSpecSecret gets the Kubernetes experiment spec secret
func (driver *KubeDriver) getExperimentSpecSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getSpecSecretName())
}

// getExperimentResultSecret gets the Kubernetes experiment result secret
func (driver *KubeDriver) getExperimentResultSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getResultSecretName())
}

// getExperimentJob gets the Kubernetes experiment job
func (driver *KubeDriver) getExperimentJob() (j *batchv1.Job, err error) {
	return driver.getJobWithRetry(driver.getExperimentJobName())
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

// create the experiment result secret
/* Example:
// # apiVersion: v1
// # kind: Secret
// # metadata:
// #   name: {{ $name }}-result
// # stringData:
// #   result.yaml: |
// #     startTime: {{ now }}
// #     numCompletedTasks: 0
// #     failure: false
// #     iter8Version: {{ .Chart.AppVersion }}
*/

// createExperimentResultSecret creates the experiment result secret
func (driver *KubeDriver) createExperimentResultSecret(r *base.ExperimentResult) error {
	job, err := driver.getExperimentJob()
	if err != nil {
		return err
	}
	// got job ...

	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	byteArray, _ := yaml.Marshal(r)
	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: driver.getResultSecretName(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: job.APIVersion,
				Kind:       job.Kind,
				Name:       job.Name,
				UID:        job.UID,
			}},
		},
		StringData: map[string]string{"result.yaml": string(byteArray)},
	}
	// formed result secret ...

	s, err := secretsClient.Create(context.Background(), &sec, metav1.CreateOptions{})
	if err != nil {
		e := errors.New("unable to create result secret")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Debug("secret data... ", s.Data)
	log.Logger.Debug("secret string data... ", s.StringData)

	// created result secret ...
	return nil
}

// updateExperimentResultSecret updates the experiment result secret
func (driver *KubeDriver) updateExperimentResultSecret(r *base.ExperimentResult) error {
	job, err := driver.getExperimentJob()
	if err != nil {
		return err
	}
	// got job ...

	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	byteArray, _ := yaml.Marshal(r)
	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: driver.getResultSecretName(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: job.APIVersion,
				Kind:       job.Kind,
				Name:       job.Name,
				UID:        job.UID,
			}},
		},
		StringData: map[string]string{"result.yaml": string(byteArray)},
	}
	// formed result secret ...

	s, err := secretsClient.Update(context.Background(), &sec, metav1.UpdateOptions{})
	if err != nil {
		e := errors.New("unable to update result secret")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Debug("secret data... ", s.Data)
	log.Logger.Debug("secret string data... ", s.StringData)

	// updated result secret ...
	return nil
}

// WriteResult writes results for a Kubernetes experiment
func (driver *KubeDriver) WriteResult(r *base.ExperimentResult) error {
	// create result secret if need be
	if sec, _ := driver.getExperimentResultSecret(); sec == nil {
		log.Logger.Info("creating experiment result secret")
		if err := driver.createExperimentResultSecret(r); err != nil {
			return err
		}
	}
	if err := driver.updateExperimentResultSecret(r); err != nil {
		return err
	}
	return nil
}

// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/upgrade.go#L69
// Upgrade a Kubernetes experiment to the next release
func (driver *KubeDriver) Upgrade(version string, chartName string, valueOpts values.Options, group string, dry bool, cpo *action.ChartPathOptions) error {
	client := action.NewUpgrade(driver.Configuration)
	client.Namespace = driver.Namespace()
	client.Version = version
	client.DryRun = dry
	client.ChartPathOptions = *cpo

	ch, vals, err := getChartAndVals(cpo, chartName, driver.EnvSettings, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and vals for %v", chartName)
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
	driver.Revision = rel.Version

	log.Logger.Info("experiment launched. Happy Iter8ing!")

	return nil
}

// Install a Kubernetes experiment
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (driver *KubeDriver) Install(version string, chartName string, valueOpts values.Options, group string, dry bool, cpo *action.ChartPathOptions) error {
	client := action.NewInstall(driver.Configuration)
	client.Namespace = driver.Namespace()
	client.Version = version
	client.DryRun = dry
	client.ChartPathOptions = *cpo
	client.ReleaseName = group

	ch, vals, err := getChartAndVals(cpo, chartName, driver.EnvSettings, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and vals for %v", chartName)
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
	driver.Revision = rel.Version

	log.Logger.Info("experiment launched. Happy Iter8ing!")

	return nil
}

// getChartAndVals gets experiment chart and its values
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func getChartAndVals(cpo *action.ChartPathOptions, chartName string, settings *cli.EnvSettings, valueOpts values.Options) (*chart.Chart, map[string]interface{}, error) {
	chartPath, err := cpo.LocateChart(chartName, settings)
	if err != nil {
		e := fmt.Errorf("unable to locate chart %v", chartName)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, nil, e
	}

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		e := fmt.Errorf("unable to merge chart values")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, nil, e
	}

	// attempt to load the chart
	ch, err := loader.Load(chartPath)
	if err != nil {
		e := fmt.Errorf("unable to load chart")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, nil, e
	}

	if err := checkIfInstallable(ch); err != nil {
		return nil, nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			man := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        chartPath,
				Keyring:          cpo.Keyring,
				SkipUpdate:       false,
				Getters:          p,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
			}
			if err := man.Update(); err != nil {
				e := fmt.Errorf("unable to update dependencies")
				log.Logger.WithStackTrace(err.Error()).Error(e)
				return nil, nil, e
			}
			// Reload the chart with the updated Chart.lock file.
			if ch, err = loader.Load(chartPath); err != nil {
				e := fmt.Errorf("failed reloading chart after dependency update")
				log.Logger.WithStackTrace(err.Error()).Error(e)
				return nil, nil, e
			}
		}
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
