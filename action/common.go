package action

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	experimentSpecPath   = "experiment.yaml"
	experimentResultPath = "result.yaml"

	specSecretSuffix   = "-spec"
	resultSecretSuffix = "-result"

	maxGetRetries    = 2
	getRetryInterval = 1 * time.Second
)

// const (
// 	SpecSecretPrefix = "experiment-"

// 	NameLabel      = "app.kubernetes.io/name"
// 	IdLabel        = "app.kubernetes.io/instance"
// 	VersionLabel   = "app.kubernetes.io/version"
// 	ComponentLabel = "app.kubernetes.io/component"
// 	CreatedByLabel = "app.kubernetes.io/created-by"
// 	AppLabel       = "iter8.tools/app"

// 	ComponentSpec   = "spec"
// 	ComponentResult = "result"
// 	ComponentJob    = "job"
// 	ComponentRbac   = "rbac"

// 	GetRetryInterval = 1 * time.Second
// )

// Build builds an experiment
func Build(eio base.ExpIO) (*base.Experiment, error) {
	e := base.Experiment{}
	var err error
	e.Tasks, err = eio.ReadSpec()
	if err != nil {
		return nil, err
	}
	e.Result, err = eio.ReadResult()
	if err != nil {
		return nil, err
	}
	return &e, nil
}

//fileIO enables reading and writing experiment spec and result files
type fileIO struct{}

// SpecFromBytes reads experiment spec from bytes
func SpecFromBytes(b []byte) (base.ExperimentSpec, error) {
	e := base.ExperimentSpec{}
	err := yaml.Unmarshal(b, &e)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment spec")
		return nil, err
	}
	return e, err
}

// ResultFromBytes reads experiment result from bytes
func ResultFromBytes(b []byte) (*base.ExperimentResult, error) {
	r := &base.ExperimentResult{}
	err := yaml.Unmarshal(b, r)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment result")
		return nil, err
	}
	return r, err
}

// ReadSpec reads experiment spec from file
func (f *fileIO) ReadSpec() (base.ExperimentSpec, error) {
	b, err := ioutil.ReadFile(experimentSpecPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, errors.New("unable to read experiment spec")
	}
	return SpecFromBytes(b)
}

// ReadResult reads experiment result from file
func (f *fileIO) ReadResult() (*base.ExperimentResult, error) {
	b, err := ioutil.ReadFile(experimentResultPath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, errors.New("unable to read experiment result")
	}
	return ResultFromBytes(b)
}

// WriteResult writes experiment result to file
func (f *fileIO) WriteResult(res *base.ExperimentResult) error {
	rBytes, _ := yaml.Marshal(res)
	err := ioutil.WriteFile(experimentResultPath, rBytes, 0664)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}

/*******************
********************

Kubernetes stuff below

********************
********************/

type ExperimentResource struct {
	cli.EnvSettings
	Group string
}

// KubeIO enables reading and writing experiment resources in Kubernetes
type KubeIO struct {
	*kubernetes.Clientset
	Group     string
	Revision  int
	Namespace string
}

func (er *ExperimentResource) newKubeIO() (*KubeIO, error) {
	// getting action config
	actionConfig := new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(er.RESTClientGetter(), er.Namespace(), helmDriver, log.Logger.Debugf); err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to get kubernetes client config")
		return nil, err
	}

	// getting last revision
	rel, err := actionConfig.Releases.Last(er.Group)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to get last experiment revision")
		return nil, err
	}

	restConfig, err := er.RESTClientGetter().ToRESTConfig()
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to get Kubernetes REST config")
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to get Kubernetes clientset")
		return nil, err
	}

	return &KubeIO{
		Clientset: clientset,
		Group:     er.Group,
		Revision:  rel.Version,
		Namespace: rel.Namespace,
	}, nil
}

func (kio *KubeIO) getSpecSecretName() string {
	return fmt.Sprintf("%v-%v-spec", kio.Group, kio.Revision)
}

func (kio *KubeIO) getResultSecretName() string {
	return fmt.Sprintf("%v-%v-result", kio.Group, kio.Revision)
}

func (kio *KubeIO) getSecretWithRetry(name string) (s *corev1.Secret, err error) {

	secretsClient := kio.CoreV1().Secrets(kio.Namespace)

	for i := 0; i < maxGetRetries; i++ {
		s, err = secretsClient.Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			return s, err
		}
		if !k8serrors.IsNotFound(err) {
			log.Logger.Errorf("unable to read secret: %s; %s\n", name, err.Error())
			return nil, err
		}
		time.Sleep(getRetryInterval)
	}
	// tried MAX_RETRIES times
	log.Logger.Errorf("experiment \"%s\" not found; unable to read secret: %s\n", name, err.Error())
	return nil, fmt.Errorf("experiment not found")
}

func (kio *KubeIO) getExperimentSpecSecret() (s *corev1.Secret, err error) {
	return kio.getSecretWithRetry(kio.getSpecSecretName())
}

func (kio *KubeIO) getExperimentResultSecret() (s *corev1.Secret, err error) {
	return kio.getSecretWithRetry(kio.getResultSecretName())
}

// read experiment spec from secret in the Kubernetes context
func (kio *KubeIO) ReadSpec() (base.ExperimentSpec, error) {
	s, err := kio.getExperimentSpecSecret()
	if err != nil {
		return nil, err
	}

	spec, ok := s.Data[experimentSpecPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment spec; spec secret has no %v field", experimentSpecPath)
		log.Logger.Error(err)
		return nil, err
	}

	return SpecFromBytes(spec)
}

// read experiment result from Kubernetes context
func (kio *KubeIO) ReadResult() (*base.ExperimentResult, error) {
	s, err := kio.getExperimentResultSecret()
	if err != nil {
		return nil, err
	}

	res, ok := s.Data[experimentResultPath]
	if !ok {
		err = fmt.Errorf("unable to extract experiment result; result secret has no %v field", experimentResultPath)
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

// write experiment result to secret in Kubernetes context
func (kio *KubeIO) WriteResult(r *base.ExperimentResult) error {
	rBytes, _ := yaml.Marshal(r)

	payload := []PayloadValue{{
		Op:    "replace",
		Path:  "/data/" + experimentResultPath,
		Value: base64.StdEncoding.EncodeToString(rBytes),
	}}
	payloadBytes, _ := json.Marshal(payload)
	secretsClient := kio.CoreV1().Secrets(kio.Namespace)

	_, err := secretsClient.Patch(context.Background(), kio.getResultSecretName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
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
