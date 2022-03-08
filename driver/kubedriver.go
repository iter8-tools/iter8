package driver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
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
	*kubernetes.Clientset
	Group    string
	Revision int
}

func (driver *KubeDriver) Init() error {
	// update revision if need be
	if driver.Revision <= 0 {
		log.Logger.Infof("fetching latest revision for experiment group %v", driver.Group)
		// getting action config
		actionConfig := new(action.Configuration)
		helmDriver := os.Getenv("HELM_DRIVER")
		if err := actionConfig.Init(driver.RESTClientGetter(), driver.Namespace(), helmDriver, log.Logger.Debugf); err != nil {
			e := errors.New("unable to get kubernetes client config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}

		// getting last revision
		rel, err := actionConfig.Releases.Last(driver.Group)
		if err != nil {
			e := fmt.Errorf("unable to get latest revision for experiment group %v", driver.Group)
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}

		driver.Revision = rel.Version
	}

	// get REST config
	restConfig, err := driver.RESTClientGetter().ToRESTConfig()
	if err != nil {
		e := errors.New("unable to get Kubernetes REST config")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	// gete clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		e := errors.New("unable to get Kubernetes clientset")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	driver.Clientset = clientset

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
	secretsClient := driver.CoreV1().Secrets(driver.Namespace())
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

	jobsClient := driver.BatchV1().Jobs(driver.Namespace())

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

	secretsClient := driver.CoreV1().Secrets(driver.Namespace())
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

	_, err = secretsClient.Create(context.Background(), &sec, metav1.CreateOptions{})
	if err != nil {
		e := errors.New("unable to create result secret")
		log.Logger.WithStackTrace(err.Error()).Error(e)
	}

	// created result secret ...
	return nil
}

// write experiment result to secret in Kubernetes context
func (driver *KubeDriver) WriteResult(r *base.ExperimentResult) error {
	// create result secret if need be
	if sec, _ := driver.getExperimentResultSecret(); sec == nil {
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

	secretsClient := driver.CoreV1().Secrets(driver.Namespace())
	_, err := secretsClient.Patch(context.Background(), driver.getResultSecretName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}

// func GetExperimentLogs(client *kubernetes.Clientset, ns string, id string) (err error) {
// 	ctx := context.Background()
// 	podList, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", IdLabel, id)})
// 	if err != nil {
// 		return err
// 	}

// 	if len(podList.Items) == 0 {
// 		return errors.New("logs not available")
// 	}

// 	for _, pod := range podList.Items {
// 		req := client.CoreV1().Pods(ns).GetLogs(pod.Name, &corev1.PodLogOptions{})
// 		logs, err := req.Stream(ctx)
// 		if err != nil {
// 			return err
// 		}
// 		buf := new(bytes.Buffer)
// 		buf.ReadFrom(logs)
// 		fmt.Println(buf.String())
// 	}
// 	return nil
// }
