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

	// auth import enables automated authentication to various hosted clouds
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
	if err := kd.initRevision(); err != nil {
		return err
	}
	return nil
}

// getLastRelease fetches the last release of an Iter8 experiment
func (kd *KubeDriver) getLastRelease() (*release.Release, error) {
	log.Logger.Debugf("fetching latest revision for experiment group %v", kd.Group)
	// getting last revision
	rel, err := kd.Configuration.Releases.Last(kd.Group)
	if err != nil {
		if helmerrors.Is(err, helmdriver.ErrReleaseNotFound) {
			log.Logger.Debugf("experiment release not found")
			return nil, nil
		}
		e := fmt.Errorf("unable to get latest revision for experiment group %v", kd.Group)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return nil, e
	}
	return rel, nil
}

// getExperimentSecretName yields the name of the experiment secret
func (kd *KubeDriver) getExperimentSecretName() string {
	return fmt.Sprintf("%v", kd.Group)
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

	b, ok := s.Data[ExperimentPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment; spec secret has no %v field", ExperimentPath)
		log.Logger.Error(err)
		return nil, err
	}

	return ExperimentFromBytes(b)
}

// formExperimentSecret creates the experiment secret using the experiment
func (kd *KubeDriver) formExperimentSecret(e *base.Experiment) (*corev1.Secret, error) {
	byteArray, err := yaml.Marshal(e)
	if err != nil {
		return nil, err
	}
	// log.Logger.Debug(string(byteArray))
	sec := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: kd.getExperimentSecretName(),
			Annotations: map[string]string{
				"iter8.tools/group": kd.Group,
			},
		},
		StringData: map[string]string{ExperimentPath: string(byteArray)},
	}
	// formed experiment secret ...
	return &sec, nil
}

// updateExperimentSecret updates the experiment secret
// as opposed to patch, update is an atomic operation
func (kd *KubeDriver) updateExperimentSecret(e *base.Experiment) error {
	if sec, err := kd.formExperimentSecret(e); err == nil {
		secretsClient := kd.Clientset.CoreV1().Secrets(kd.Namespace())
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

	// write to secret
	if err := kd.updateExperimentSecret(exp); err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment")
		return errors.New("unable to write experiment")
	}
	return nil
}

// GetRevision gets the experiment revision
func (kd *KubeDriver) GetRevision() int {
	return kd.revision
}

// writeManifest writes the Kubernetes experiment manifest to a local file
func writeManifest(rel *release.Release) error {
	err := os.WriteFile(ManifestFile, []byte(rel.Manifest), 0600)
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
func (kd *KubeDriver) upgrade(chartPathOptions action.ChartPathOptions, chartName string, valueOpts values.Options, group string, dry bool) error {
	client := action.NewUpgrade(kd.Configuration)
	client.Namespace = kd.Namespace()
	client.DryRun = dry

	// copy chartpathoptions
	// client.ChartPathOptions.CaFile = chartPathOptions.CaFile
	// client.ChartPathOptions.CertFile = chartPathOptions.CertFile
	// client.ChartPathOptions.KeyFile = chartPathOptions.KeyFile
	// client.ChartPathOptions.InsecureSkipTLSverify = chartPathOptions.InsecureSkipTLSverify
	// client.ChartPathOptions.Keyring = chartPathOptions.Keyring
	// client.ChartPathOptions.Password = chartPathOptions.Password
	// client.ChartPathOptions.PassCredentialsAll = chartPathOptions.PassCredentialsAll
	client.ChartPathOptions.RepoURL = chartPathOptions.RepoURL
	// client.ChartPathOptions.Username = chartPathOptions.Username
	// client.ChartPathOptions.Verify = chartPathOptions.Verify
	client.ChartPathOptions.Version = chartPathOptions.Version

	cp, err := client.ChartPathOptions.LocateChart(chartName, kd.EnvSettings)
	if err != nil {
		log.Logger.Error(err)
		return err
	}

	chartRequested, vals, err := kd.getChartAndVals(cp, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and value for %v", cp)
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
		log.Logger.Warnf("experiment for group %s has been cancelled.\n", group)
		cancel()
	}()

	rel, err := client.RunWithContext(ctx, group, chartRequested, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	return kd.updateRevision(rel, dry)
}

// install a Kubernetes experiment
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (kd *KubeDriver) install(chartPathOptions action.ChartPathOptions, chartName string, valueOpts values.Options, group string, dry bool) error {

	// buf := new(bytes.Buffer)
	client := action.NewInstall(kd.Configuration)
	client.Namespace = kd.Namespace()
	client.RepoURL = chartPathOptions.RepoURL
	client.DryRun = dry
	client.ReleaseName = group

	// copy chartPathOptions to client
	// client.ChartPathOptions.CaFile = chartPathOptions.CaFile
	// client.ChartPathOptions.CertFile = chartPathOptions.CertFile
	// client.ChartPathOptions.KeyFile = chartPathOptions.KeyFile
	// client.ChartPathOptions.InsecureSkipTLSverify = chartPathOptions.InsecureSkipTLSverify
	// client.ChartPathOptions.Keyring = chartPathOptions.Keyring
	// client.ChartPathOptions.Password = chartPathOptions.Password
	// client.ChartPathOptions.PassCredentialsAll = chartPathOptions.PassCredentialsAll
	client.ChartPathOptions.RepoURL = chartPathOptions.RepoURL
	// client.ChartPathOptions.Username = chartPathOptions.Username
	// client.ChartPathOptions.Verify = chartPathOptions.Verify
	client.ChartPathOptions.Version = chartPathOptions.Version

	cp, err := client.ChartPathOptions.LocateChart(chartName, kd.EnvSettings)
	if err != nil {
		log.Logger.Error(err)
		return err
	}

	chartRequested, vals, err := kd.getChartAndVals(cp, valueOpts)
	if err != nil {
		e := fmt.Errorf("unable to get chart and value for %v", cp)
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
		log.Logger.Warnf("experiment for group %s has been cancelled.\n", group)
		cancel()
	}()

	rel, err := client.RunWithContext(ctx, chartRequested, vals)
	if err != nil {
		e := fmt.Errorf("experiment launch failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	return kd.updateRevision(rel, dry)
}

func (kd *KubeDriver) updateRevision(rel *release.Release, dry bool) error {
	// upgrading revision info
	kd.revision = rel.Version

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
func (kd *KubeDriver) Launch(chartPathOptions action.ChartPathOptions, chartName string, valueOpts values.Options, group string, dry bool) error {
	if kd.revision <= 0 {
		return kd.install(chartPathOptions, chartName, valueOpts, group, dry)
	}
	return kd.upgrade(chartPathOptions, chartName, valueOpts, group, dry)
}

// Delete a Kubernetes experiment group
func (kd *KubeDriver) Delete() error {
	client := action.NewUninstall(kd.Configuration)
	_, err := client.Run(kd.Group)
	if err != nil {
		e := fmt.Errorf("deletion of experiment group %v failed", kd.Group)
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	log.Logger.Infof("experiment group %v deleted", kd.Group)
	return nil
}

// getChartAndVals gets experiment chart and its values
// Credit: the logic for this function is sourced from Helm
// https://github.com/helm/helm/blob/8ab18f7567cedffdfa5ba4d7f6abfb58efc313f8/cmd/helm/install.go#L177
func (kd *KubeDriver) getChartAndVals(chartDir string, valueOpts values.Options) (*chart.Chart, map[string]interface{}, error) {
	// form chart values
	p := getter.All(kd.EnvSettings)
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
func (kd *KubeDriver) GetExperimentLogs() (string, error) {
	podsClient := kd.Clientset.CoreV1().Pods(kd.Namespace())
	pods, err := podsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("iter8.tools/group=%v", kd.Group),
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
			e := fmt.Errorf("error in opening log stream: %e", err)
			log.Logger.Error(e)
			return "", e
		}

		defer func() {
			_ = podLogs.Close()
		}()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			e := fmt.Errorf("error in copy information from podLogs to buf: %e", err)
			log.Logger.Error(e)
			return "", e
		}
		str := buf.String()
		lgs[i] = str
	}
	return strings.Join(lgs, "\n***\n"), nil
}
