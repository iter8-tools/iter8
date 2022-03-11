package driver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	maxGetRetries    = 2
	getRetryInterval = 1 * time.Second
)

type KubeDriver struct {
	*cli.EnvSettings
	Clientset kubernetes.Interface
	Group     string
	Revision  int
}

func (driver *KubeDriver) getHelmConfig() (*action.Configuration, error) {
	// getting kube config
	actionConfig := new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(driver.RESTClientGetter(), driver.Namespace(), helmDriver, log.Logger.Debugf); err != nil {
		e := errors.New("unable to get Helm client config")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	return actionConfig, nil
}

func (driver *KubeDriver) getLastRelease() (*release.Release, error) {
	log.Logger.Infof("fetching latest revision for experiment group %v", driver.Group)
	// get kube config
	actionConfig, err := driver.getHelmConfig()
	if err != nil {
		return nil, err
	}
	// getting last revision
	rel, err := actionConfig.Releases.Last(driver.Group)
	if err != nil {
		e := fmt.Errorf("unable to get latest revision for experiment group %v", driver.Group)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	return rel, nil
}

func (driver *KubeDriver) Init() error {
	// update revision to latest, if none is specified
	if driver.Revision <= 0 {
		if rel, err := driver.getLastRelease(); err == nil {
			driver.Revision = rel.Version
		} else {
			return err
		}
	}

	if driver.Clientset == nil { // initialize
		// get REST config
		restConfig, err := driver.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		driver.Clientset = clientset
	}

	return nil
}

func (driver *KubeDriver) getSpecSecretName() string {
	return fmt.Sprintf("%v-%v-spec", driver.Group, driver.Revision)
}

func (driver *KubeDriver) getResultSecretName() string {
	return fmt.Sprintf("%v-%v-result", driver.Group, driver.Revision)
}

func (driver *KubeDriver) getExperimentJobName() string {
	return fmt.Sprintf("%v-%v-job", driver.Group, driver.Revision)
}

func (driver *KubeDriver) getSecretWithRetry(name string) (s *corev1.Secret, err error) {
	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	for i := 0; i < maxGetRetries; i++ {
		s, err = secretsClient.Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return s, err
		}
		if !k8serrors.IsNotFound(err) {
			log.Logger.Errorf("unable to get secret: %s; %s\n", name, err.Error())
		}
		time.Sleep(getRetryInterval)
	}
	e := fmt.Errorf("unable to get secret %v", name)
	log.Logger.Error(e)
	return nil, e
}

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

func (driver *KubeDriver) getExperimentSpecSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getSpecSecretName())
}

func (driver *KubeDriver) getExperimentResultSecret() (s *corev1.Secret, err error) {
	return driver.getSecretWithRetry(driver.getResultSecretName())
}

func (driver *KubeDriver) getExperimentJob() (j *batchv1.Job, err error) {
	return driver.getJobWithRetry(driver.getExperimentJobName())
}

// read experiment spec from secret in the Kubernetes context
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

// read experiment result from Kubernetes context
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

type PayloadValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
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
func (driver *KubeDriver) createExperimentResultSecret(r *base.ExperimentResult) error {
	job, err := driver.getExperimentJob()
	if err != nil {
		return err
	}
	// get job ...

	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	rYaml, _ := yaml.Marshal(r)
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
		StringData: map[string]string{"result.yaml": string(rYaml)},
	}
	// formed result secret ...

	s, err := secretsClient.Create(context.Background(), &sec, metav1.CreateOptions{})
	if err != nil {
		e := errors.New("unable to create result secret")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Info("secret data... ", s.Data)
	log.Logger.Info("secret string data... ", s.StringData)

	// created result secret ...
	return nil
}

// write experiment result to secret in Kubernetes context
func (driver *KubeDriver) WriteResult(r *base.ExperimentResult) error {
	// create result secret if need be
	if sec, _ := driver.getExperimentResultSecret(); sec == nil {
		log.Logger.Info("creating experiment result secret")
		if err := driver.createExperimentResultSecret(r); err != nil {
			return err
		}
	}
	// result secret exists at this point ...

	rBytes, _ := yaml.Marshal(r)

	payload := []PayloadValue{{
		Op:    "replace",
		Path:  "/data/" + ExperimentResultPath,
		Value: base64.StdEncoding.EncodeToString(rBytes),
	}}
	payloadBytes, _ := json.Marshal(payload)

	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
	_, err := secretsClient.Patch(context.Background(), driver.getResultSecretName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}

// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/upgrade.go#L69
func (driver *KubeDriver) Upgrade(version string, chartName string, valueOpts values.Options, group string, dry bool, cpo *action.ChartPathOptions) error {
	cfg, err := driver.getHelmConfig()
	if err != nil {
		return err
	}
	client := action.NewUpgrade(cfg)
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

	_, err = client.RunWithContext(ctx, group, ch, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Info("Experiment launched. Happy Iter8ing!")

	return nil
}

// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (driver *KubeDriver) Install(version string, chartName string, valueOpts values.Options, group string, dry bool, cpo *action.ChartPathOptions) error {
	cfg, err := driver.getHelmConfig()
	if err != nil {
		return err
	}
	client := action.NewInstall(cfg)
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

	_, err = client.RunWithContext(ctx, ch, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	log.Logger.Info("Experiment launched. Happy Iter8ing!")

	return nil
}

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

	out := os.Stdout

	if err := checkIfInstallable(ch); err != nil {
		return nil, nil, err
	}

	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			man := &downloader.Manager{
				Out:              out,
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

// Credit: this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L270
//
// checkIfInstallable validates if a chart can be installed
// Only application chart type is installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	e := fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
	log.Logger.Error(e)
	return e
}

func (driver *KubeDriver) getJobName() string {
	return fmt.Sprintf("%v-%v-job", driver.Group, driver.Revision)
}

func (driver *KubeDriver) GetExperimentLogs() (string, error) {
	podsClient := driver.Clientset.CoreV1().Pods(driver.Namespace())
	pods, err := podsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%v", driver.getJobName()),
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
