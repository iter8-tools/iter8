package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"
	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
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

func GetExperimentSecret(client *kubernetes.Clientset, ns string, nm string) (experiment *corev1.Secret, err error) {
	ctx := context.Background()

	// A name is provided; get this experiment, if it exists
	if len(nm) != 0 {
		experiment, err = client.CoreV1().Secrets(ns).Get(ctx, nm, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, fmt.Errorf("experiment \"%s\" not found", nm)
			}
		}
		// verify that the job corresponds to an experiment
		if experiment != nil && !isExperiment(*experiment) {
			return nil, fmt.Errorf("experiment \"%s\" not found", nm)
		}

		return experiment, err
	}

	// There is no explict experiment name provided.
	// Get a list of all experiments.
	// Then select the one with the most recent create time.
	experiments, err := GetExperimentSecrets(client, ns)
	if err != nil {
		return experiment, err
	}

	// no experiments
	if len(experiments) == 0 {
		return experiment, errors.New("no experiments found")
	}

	for _, job := range experiments {
		if experiment == nil {
			experiment = &job
			continue
		}
		if job.ObjectMeta.CreationTimestamp.Time.After(experiment.ObjectMeta.CreationTimestamp.Time) {
			experiment = &job
		}
	}
	return experiment, nil
}

func GetExperimentSecrets(client *kubernetes.Clientset, ns string) (experiments []corev1.Secret, err error) {
	secrets, err := client.CoreV1().Secrets(ns).List(
		context.Background(), metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=iter8,app.kubernetes.io/component=spec",
		})
	if err != nil {
		return experiments, err
	}

	return jobListToExperimentJobList(*secrets), err
}

func jobListToExperimentJobList(secrets corev1.SecretList) (result []corev1.Secret) {
	for _, secret := range secrets.Items {
		if isExperiment(secret) {
			result = append(result, secret)
		}
	}
	return result
}

func isExperiment(e corev1.Secret) bool {
	return !strings.HasSuffix(e.GetName(), "-result")
}

//KubernetesExpIO enables reading and writing through files
type KubernetesExpIO struct {
	Client    *kubernetes.Clientset
	Namespace string
	Name      string
}

// read experiment spec from secret in the Kubernetes context
func (f *KubernetesExpIO) ReadSpec() ([]base.TaskSpec, error) {

	s, err := f.Client.CoreV1().Secrets(f.Namespace).Get(context.Background(), f.Name, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return nil, fmt.Errorf("experiment \"%s\" not found", f.Name)
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
	s, err := f.Client.CoreV1().Secrets(f.Namespace).Get(context.Background(), resultSecretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment result")
		return nil, fmt.Errorf("experiment \"%s\" not found", f.Name)
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
