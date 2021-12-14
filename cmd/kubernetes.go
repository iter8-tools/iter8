package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/basecli"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

const (
	SpecSecretPrefix = "experiment-"

	NameLabel      = "app.kubernetes.io/name"
	IdLabel        = "app.kubernetes.io/instance"
	VersionLabel   = "app.kubernetes.io/version"
	ComponentLabel = "app.kubernetes.io/component"
	CreatedByLabel = "app.kubernetes.io/created-by"
	AppLabel       = "iter8.tools/app"

	ComponentSpec   = "spec"
	ComponentResult = "result"
	ComponentJob    = "job"
	ComponentRbac   = "rbac"

	MaxGetRetries    = 2
	GetRetryInterval = 1 * time.Second
)

func GetClient(cf *genericclioptions.ConfigFlags) (*kubernetes.Clientset, error) {
	restConfig, err := cf.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func GetExperimentLogs(client *kubernetes.Clientset, ns string, id string) (err error) {
	ctx := context.Background()
	podList, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", IdLabel, id)})
	if err != nil {
		return err
	}

	if len(podList.Items) == 0 {
		return errors.New("logs not available")
	}

	for _, pod := range podList.Items {
		req := client.CoreV1().Pods(ns).GetLogs(pod.Name, &corev1.PodLogOptions{})
		logs, err := req.Stream(ctx)
		if err != nil {
			return err
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(logs)
		fmt.Println(buf.String())
	}
	return nil
}

func getSecretWithRetry(client *kubernetes.Clientset, ns string, nm string) (s *corev1.Secret, err error) {
	for i := 0; i < MaxGetRetries; i++ {
		s, err = client.CoreV1().Secrets(ns).Get(context.Background(), nm, metav1.GetOptions{})
		if err == nil {
			return s, err
		}
		if !k8serrors.IsNotFound(err) {
			log.Logger.Errorf("unable to read secret: %s; %s\n", nm, err.Error())
			return nil, err
		}
		time.Sleep(GetRetryInterval)
	}
	// tried MAX_RETRIES times
	log.Logger.Errorf("experiment \"%s\" not fouund; unable to read secret: %s\n", nm, err.Error())
	return nil, fmt.Errorf("experiment not found")
}

func GetExperimentSecret(client *kubernetes.Clientset, ns string, id string) (s *corev1.Secret, err error) {
	// An id is provided; get this experiment, if it exists
	if len(id) != 0 {
		nm := SpecSecretPrefix + id
		s, err = getSecretWithRetry(client, ns, nm)
		// verify that the secret (if found) corresponds to an experiment
		if s != nil && err == nil && !isExperiment(*s) {
			return nil, fmt.Errorf("experiment not found")
		}
		return s, err
	}

	// There is no explict experiment name provided.
	// Get a list of all experiments.
	// Then select the one with the most recent create time.
	experimentSecrets, err := GetExperimentSecrets(client, ns)
	if err != nil {
		return s, err
	}

	// no experiments
	if len(experimentSecrets) == 0 {
		return s, errors.New("no experiments found")
	}

	mostRecent := experimentSecrets[len(experimentSecrets)-1]
	return &mostRecent, nil
}

func GetExperimentSecrets(client *kubernetes.Clientset, ns string) (experimentSecrets []corev1.Secret, err error) {
	secrets, err := client.CoreV1().Secrets(ns).List(
		context.Background(), metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=iter8,app.kubernetes.io/component=spec",
		})
	if err != nil {
		return experimentSecrets, err
	}

	return secretListToOrderedExperimentSecretList(*secrets), err
}

func secretListToOrderedExperimentSecretList(secrets corev1.SecretList) (results []corev1.Secret) {
	results = []corev1.Secret{}
	for _, secret := range secrets.Items {
		if !isExperiment(secret) {
			continue
		}
		index := len(results)
		for j, r := range results {
			if !secret.CreationTimestamp.Time.After(r.CreationTimestamp.Time) {
				index = j
				break
			}
		}
		if index < len(results) {
			results = append(results[:index+1], results[index:]...)
			results[index] = secret
		} else {
			results = append(results, secret)
		}
	}
	return results
}

func isExperiment(s corev1.Secret) bool {
	component, ok := s.Labels[ComponentLabel]
	if !ok {
		return false
	}
	return component == ComponentSpec
}

//KubernetesExpIO enables reading and writing through files
type KubernetesExpIO struct {
	Client    *kubernetes.Clientset
	Namespace string
	Name      string
}

// read experiment spec from secret in the Kubernetes context
func (f *KubernetesExpIO) ReadSpec() ([]base.TaskSpec, error) {

	s, err := getSecretWithRetry(f.Client, f.Namespace, f.Name)
	if err != nil {
		return nil, err
	}

	exp, ok := s.Data["experiment"]
	if !ok {
		log.Logger.Error("unable to read experiment spec; spec secret has no experiment field")
		return nil, fmt.Errorf("experiment \"%s\" not found", f.Name)
	}

	return basecli.SpecFromBytes(exp)
}

// read experiment result from Kubernetes context
func (f *KubernetesExpIO) ReadResult() (*base.ExperimentResult, error) {
	resultSecretName := f.Name + "-result"
	s, err := getSecretWithRetry(f.Client, f.Namespace, resultSecretName)
	if err != nil {
		return nil, err
	}

	r, ok := s.Data["result"]
	if !ok {
		log.Logger.Error("unable to read experiment spec; result secret has no data field")
		return nil, fmt.Errorf("experiment \"%s\" result not found", f.Name)
	}

	return basecli.ResultFromBytes(r)
}

type PayloadValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// write experiment result to secret in Kubernetes context
func (f *KubernetesExpIO) WriteResult(r *basecli.Experiment) error {
	rBytes, _ := yaml.Marshal(r.Result)

	resultSecretName := f.Name + "-result"

	payload := []PayloadValue{{
		Op:    "replace",
		Path:  "/data/result",
		Value: base64.StdEncoding.EncodeToString(rBytes),
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, err := f.Client.CoreV1().Secrets(f.Namespace).Patch(context.Background(), resultSecretName, types.JSONPatchType, payloadBytes, metav1.PatchOptions{})

	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment result")
		return err
	}
	return err
}

const (
	ExperimentId            = "experiment-id"
	ExperimentIdShort       = "e"
	ExperimentIdDescription = "remote experiment identifier; if not specified, the most recent experiment is used"
)

func (o *K8sExperimentOptions) addExperimentIdOption(p *pflag.FlagSet) {
	// Add options
	p.StringVarP(&o.experimentId, ExperimentId, ExperimentIdShort, "", ExperimentIdDescription)
}

type K8sExperimentOptions struct {
	ConfigFlags  *genericclioptions.ConfigFlags
	namespace    string
	client       *kubernetes.Clientset
	experimentId string
	expIO        *KubernetesExpIO
	experiment   *basecli.Experiment
}

func newK8sExperimentOptions() *K8sExperimentOptions {
	return &K8sExperimentOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag(),
	}
}

func (o *K8sExperimentOptions) initK8sExperiment(withResult bool) (err error) {
	o.namespace, _, err = o.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.client, err = GetClient(o.ConfigFlags)
	if err != nil {
		return err
	}

	if len(o.experimentId) == 0 {
		s, err := GetExperimentSecret(o.client, o.namespace, o.experimentId)
		if err != nil {
			return err
		}
		o.experimentId = s.Labels[IdLabel]
	}

	o.expIO = &KubernetesExpIO{
		Client:    o.client,
		Namespace: o.namespace,
		Name:      SpecSecretPrefix + o.experimentId,
	}

	o.experiment, err = basecli.Build(withResult, o.expIO)

	return err
}
