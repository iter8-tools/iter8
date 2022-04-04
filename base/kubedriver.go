package base

import (
	// "bytes"
	"context"
	"errors"
	"fmt"

	// "io"
	"strings"
	// "time"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/iter8/base/log"

	// "sigs.k8s.io/yaml"

	// batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	// k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// const (
// 	// maxGetRetries is the number of tries to retry while fetching Kubernetes objects
// 	maxGetRetries = 2
// 	// getRetryInterval is the duration between retrials
// 	getRetryInterval = 1 * time.Second
// )

// KubeDriver embeds Kube configuration, and
// enables interaction with a Kubernetes cluster through Kube APIs
type KubeDriver struct {
	// EnvSettings provides generic Kubernetes options
	// *cli.EnvSettings
	*EnvSettings
	// Clientset enables interaction with a Kubernetes cluster
	Clientset kubernetes.Interface
	// Group is the experiment group
	Group string
	// Revision is the revision of the experiment
	Revision      int
	RestConfig    *rest.Config
	GetObjectFunc GetObjectFuncType
}

type GetObjectFuncType func(*KubeDriver, *corev1.ObjectReference) (*unstructured.Unstructured, error)

// NewKubeDriver creates and returns a new KubeDriver
func NewKubeDriver(s *EnvSettings) *KubeDriver {
	kd := &KubeDriver{
		EnvSettings:   s,
		Group:         DefaultExperimentGroup,
		Clientset:     nil,
		RestConfig:    nil,
		GetObjectFunc: GetRealObject,
	}
	return kd
}

// initKube initializes the Kubernetes clientset
func (kd *KubeDriver) initKube() (err error) {
	if kd.Clientset == nil {
		// get REST config
		kd.RestConfig, err = kd.EnvSettings.RESTClientGetter().ToRESTConfig()
		if err != nil {
			e := errors.New("unable to get Kubernetes REST config")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
		// get clientset
		kd.Clientset, err = kubernetes.NewForConfig(kd.RestConfig)
		if err != nil {
			e := errors.New("unable to get Kubernetes clientset")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return e
		}
	}

	return nil
}

func GetFakeObject(kd *KubeDriver, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	if strings.EqualFold("pod", objRef.Kind) {
		pod, err := kd.Clientset.CoreV1().Pods(objRef.Namespace).Get(context.Background(), objRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		if err != nil {
			return nil, err
		}
		return &unstructured.Unstructured{Object: obj}, nil

	}
	return nil, fmt.Errorf("resource %s not supported", objRef.Kind)
}

// getObject finds the object referenced by objRef using the client config restConfig
// uses the dynamic client; ie, retuns an unstructured object
// based on https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func GetRealObject(kd *KubeDriver, objRef *corev1.ObjectReference) (*unstructured.Unstructured, error) {
	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(kd.RestConfig)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(kd.RestConfig)
	if err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(objRef.APIVersion, objRef.Kind)

	// 3. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	// 4. Obtain REST interface for the GVR
	namespace := objRef.Namespace // recall that we always set this
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	obj, err := dr.Get(context.Background(), objRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Init initializes the KubeDriver
func (driver *KubeDriver) Init() error {
	if err := driver.initKube(); err != nil {
		return err
	}
	return nil
}

// // getSpecSecretName yields the name of the experiment spec secret
// func (driver *KubeDriver) getSpecSecretName() string {
// 	return fmt.Sprintf("%v-%v-spec", driver.Group, driver.Revision)
// }

// // getResultSecretName yields the name of the experiment result secret
// func (driver *KubeDriver) getResultSecretName() string {
// 	return fmt.Sprintf("%v-%v-result", driver.Group, driver.Revision)
// }

// // getExperimentJobName yields the name of the experiment job
// func (driver *KubeDriver) getExperimentJobName() string {
// 	return fmt.Sprintf("%v-%v-job", driver.Group, driver.Revision)
// }

// // getSecretWithRetry attempts to get a Kubernetes secret with retry
// func (driver *KubeDriver) getSecretWithRetry(name string) (s *corev1.Secret, err error) {
// 	secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
// 	for i := 0; i < maxGetRetries; i++ {
// 		s, err = secretsClient.Get(context.Background(), name, metav1.GetOptions{})
// 		if err == nil {
// 			return s, err
// 		}
// 		if !k8serrors.IsNotFound(err) {
// 			log.Logger.Warningf("unable to get secret: %s; %s\n", name, err.Error())
// 		}
// 		time.Sleep(getRetryInterval)
// 	}
// 	e := fmt.Errorf("unable to get secret %v", name)
// 	log.Logger.Warning(e)
// 	return nil, e
// }

// // getJobWithRetry attempts to get a Kubernetes job with retry
// func (driver *KubeDriver) getJobWithRetry(name string) (*batchv1.Job, error) {

// 	jobsClient := driver.Clientset.BatchV1().Jobs(driver.Namespace())

// 	for i := 0; i < maxGetRetries; i++ {
// 		j, err := jobsClient.Get(context.Background(), name, metav1.GetOptions{})
// 		if err == nil {
// 			return j, err
// 		}
// 		if !k8serrors.IsNotFound(err) {
// 			log.Logger.Warningf("unable to get job: %s; %s\n", name, err.Error())
// 		}
// 		time.Sleep(getRetryInterval)
// 	}
// 	e := fmt.Errorf("unable to get job %v", name)
// 	log.Logger.Error(e)
// 	return nil, e
// }

// // getExperimentSpecSecret gets the Kubernetes experiment spec secret
// func (driver *KubeDriver) getExperimentSpecSecret() (s *corev1.Secret, err error) {
// 	return driver.getSecretWithRetry(driver.getSpecSecretName())
// }

// // getExperimentResultSecret gets the Kubernetes experiment result secret
// func (driver *KubeDriver) getExperimentResultSecret() (s *corev1.Secret, err error) {
// 	return driver.getSecretWithRetry(driver.getResultSecretName())
// }

// // getExperimentJob gets the Kubernetes experiment job
// func (driver *KubeDriver) getExperimentJob() (j *batchv1.Job, err error) {
// 	return driver.getJobWithRetry(driver.getExperimentJobName())
// }

// // ReadSpec creates an ExperimentSpec struct for a Kubernetes experiment
// func (driver *KubeDriver) ReadSpec() (ExperimentSpec, error) {
// 	s, err := driver.getExperimentSpecSecret()
// 	if err != nil {
// 		return nil, err
// 	}

// 	spec, ok := s.Data[ExperimentSpecPath]
// 	if !ok {
// 		err = fmt.Errorf("unable to extract experiment spec; spec secret has no %v field", ExperimentSpecPath)
// 		log.Logger.Error(err)
// 		return nil, err
// 	}

// 	return SpecFromBytes(spec)
// }

// // ReadResult creates an ExperimentResult struct for a Kubernetes experiment
// func (driver *KubeDriver) ReadResult() (*ExperimentResult, error) {
// 	s, err := driver.getExperimentResultSecret()
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, ok := s.Data[ExperimentResultPath]
// 	if !ok {
// 		err = fmt.Errorf("unable to extract experiment result; result secret has no %v field", ExperimentResultPath)
// 		log.Logger.Error(err)
// 		return nil, err
// 	}

// 	return ResultFromBytes(res)
// }

// // PayloadValue is used to patch Kubernetes resources
// type PayloadValue struct {
// 	// Op indicates the type of patch
// 	Op string `json:"op"`
// 	// Path is the JSON field path
// 	Path string `json:"path"`
// 	// Value is the value of the field
// 	Value string `json:"value"`
// }

// // create the experiment result secret
// /* Example:
// // # apiVersion: v1
// // # kind: Secret
// // # metadata:
// // #   name: {{ $name }}-result
// // # stringData:
// // #   result.yaml: |
// // #     startTime: {{ now }}
// // #     numCompletedTasks: 0
// // #     failure: false
// // #     iter8Version: {{ .Chart.AppVersion }}
// */

// // formResultSecret creates the result secret using the result
// func (driver *KubeDriver) formResultSecret(r *ExperimentResult) (*corev1.Secret, error) {
// 	job, err := driver.getExperimentJob()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// got job ...

// 	byteArray, _ := yaml.Marshal(r)
// 	sec := corev1.Secret{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: driver.getResultSecretName(),
// 			OwnerReferences: []metav1.OwnerReference{{
// 				APIVersion: job.APIVersion,
// 				Kind:       job.Kind,
// 				Name:       job.Name,
// 				UID:        job.UID,
// 			}},
// 		},
// 		StringData: map[string]string{"result.yaml": string(byteArray)},
// 	}
// 	// formed result secret ...
// 	return &sec, nil
// }

// // createExperimentResultSecret creates the experiment result secret
// func (driver *KubeDriver) createExperimentResultSecret(r *ExperimentResult) error {
// 	if sec, err := driver.formResultSecret(r); err == nil {
// 		secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
// 		_, e := secretsClient.Create(context.Background(), sec, metav1.CreateOptions{})
// 		if e != nil {
// 			e := errors.New("unable to create result secret")
// 			log.Logger.WithStackTrace(err.Error()).Error(e)
// 			return e
// 		} else {
// 			return nil
// 		}
// 	} else {
// 		return err
// 	}
// }

// // updateExperimentResultSecret updates the experiment result secret
// // as opposed to patch, update is an atomic operation
// // eventually, this code will leverage conflict management like the following:
// // https://github.com/kubernetes/client-go/blob/3ac142e26bc61901240b68cc2c39561d2e6f672a/examples/create-update-delete-deployment/main.go#L118
// func (driver *KubeDriver) updateExperimentResultSecret(r *ExperimentResult) error {
// 	if sec, err := driver.formResultSecret(r); err == nil {
// 		secretsClient := driver.Clientset.CoreV1().Secrets(driver.Namespace())
// 		_, e := secretsClient.Update(context.Background(), sec, metav1.UpdateOptions{})
// 		if e != nil {
// 			e := errors.New("unable to update result secret")
// 			log.Logger.WithStackTrace(err.Error()).Error(e)
// 			return e
// 		} else {
// 			return nil
// 		}
// 	} else {
// 		return err
// 	}
// }

// // WriteResult writes results for a Kubernetes experiment
// func (driver *KubeDriver) WriteResult(r *ExperimentResult) error {
// 	// create result secret if need be
// 	if sec, _ := driver.getExperimentResultSecret(); sec == nil {
// 		log.Logger.Info("creating experiment result secret")
// 		if err := driver.createExperimentResultSecret(r); err != nil {
// 			return err
// 		}
// 	}
// 	if err := driver.updateExperimentResultSecret(r); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // GetExperimentLogs gets logs for a Kubernetes experiment
// func (driver *KubeDriver) GetExperimentLogs() (string, error) {
// 	podsClient := driver.Clientset.CoreV1().Pods(driver.Namespace())
// 	pods, err := podsClient.List(context.TODO(), metav1.ListOptions{
// 		LabelSelector: fmt.Sprintf("job-name=%v", driver.getExperimentJobName()),
// 	})
// 	if err != nil {
// 		e := errors.New("unable to get experiment pod(s)")
// 		log.Logger.Error(e)
// 		return "", e
// 	}
// 	lgs := make([]string, len(pods.Items))
// 	for i, p := range pods.Items {
// 		req := podsClient.GetLogs(p.Name, &corev1.PodLogOptions{})
// 		podLogs, err := req.Stream(context.TODO())
// 		if err != nil {
// 			e := errors.New("error in opening log stream")
// 			log.Logger.Error(e)
// 			return "", e
// 		}

// 		defer podLogs.Close()
// 		buf := new(bytes.Buffer)
// 		_, err = io.Copy(buf, podLogs)
// 		if err != nil {
// 			e := errors.New("error in copy information from podLogs to buf")
// 			log.Logger.Error(e)
// 			return "", e
// 		}
// 		str := buf.String()
// 		lgs[i] = str
// 	}
// 	return strings.Join(lgs, "\n***\n"), nil
// }

type FakeRESTMapper struct{}

func (m *FakeRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

// KindsFor takes a partial resource and returns the list of potential kinds in priority order
func (m *FakeRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	return []schema.GroupVersionKind{}, nil
}

// ResourceFor takes a partial resource and returns the single match.  Returns an error if there are multiple matches
func (m *FakeRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	return schema.GroupVersionResource{}, nil
}

// ResourcesFor takes a partial resource and returns the list of potential resource in priority order
func (m *FakeRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	return []schema.GroupVersionResource{}, nil
}

// RESTMapping identifies a preferred resource mapping for the provided group kind.
func (m *FakeRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	return nil, nil
}

// RESTMappings returns all resource mappings for the provided group kind if no
// version search is provided. Otherwise identifies a preferred resource mapping for
// the provided version(s).
func (m *FakeRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	return []*meta.RESTMapping{}, nil
}

func (m *FakeRESTMapper) ResourceSingularizer(resource string) (singular string, err error) {
	return "", nil
}

var _ meta.RESTMapper = &FakeRESTMapper{}
