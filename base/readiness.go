package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	// "regexp"
	"strings"
	"time"

	log "github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	// "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

// TBD Are RFC 1123 names sufficient? It seems like different objects have different rewquirements
// see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

const (
	// ReadinessTaskName is the task name
	ReadinessTaskName = "k8s-objects-ready"

	// // labelFmt is a regular expression to be used to validate RFC 1123 labels (object names)
	// // see: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// labelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

	// // maxLabelLength is the maximum length of RFC 1123 labels (object names)
	// // see: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// maxLabelLength int = 63

	// defaultTimeout is default timeout for readiness command
	defaultTimeout = "10s"
)

// // regex object
// var labelRegexp = regexp.MustCompile("^" + labelFmt + "$")

// ReadinessInputs identifies the K8s object to test for existence and
// the (optional) condition that should be tested (succeeds if true).
type readinessInputs struct {
	// Kind of the object. Specified in the TYPE[.VERSION][.GROUP] format used by `kubectl`
	// See https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#get
	Kind string `json:"kind" yaml:"kind"`
	// Namespace of the object. Optional. If left unspecified, this will be defaulted to the namespace of the experiment
	Namespace *string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Name of the object
	Name string `json:"name" yaml:"name"`
	// Condition is label of condition to check for value of "True"
	Condition *string `json:"condition" yaml:"condition"`
	// Timeout is maximum time spent trying to find object and check condition
	Timeout *string `json:"timeout" yaml:"timeout"`
}

// ReadinessTask checks existence and readiness of specified resources
type readinessTask struct {
	TaskMeta
	With readinessInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the readiness task
func (t *readinessTask) initializeDefaults() {
	if t.With.Timeout == nil {
		t.With.Timeout = StringPointer(defaultTimeout)
	}
}

// validateInputs validates task inputs
func (t *readinessTask) validateInputs() error {
	// // validate that name is a RFC 1123 label
	// errs := validation.IsDNS1123Label(t.With.Name)
	// if len(errs) > 0 {
	// 	return errors.New(errs[0])
	// }

	// validate that timeout is parsable
	if t.With.Timeout != nil {
		_, err := time.ParseDuration(*t.With.Timeout)
		if err != nil {
			return errors.New("invalid format for timeout")
		}
	}

	return nil
}

// run executes the task
func (t *readinessTask) run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	// get kubeconfig from whatever is available
	// works if in cluster or out of cluster
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// get client config (*rest.Config)
	restConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return err
	}

	// set Namespace (from context) if not already set
	if t.With.Namespace == nil {
		ns, _, err := kubeconfig.Namespace()
		if err != nil {
			return err
		}
		t.With.Namespace = StringPointer(ns)
	}

	timeout, err := time.ParseDuration(*t.With.Timeout)
	if err != nil {
		return err
	}
	log.Logger.Trace("duration is ", timeout)

	// check for object and condition
	// repeat until time out
	interval := 1 * time.Second
	err = retry.OnError(
		wait.Backoff{
			Steps:    int(timeout / interval),
			Cap:      timeout,
			Duration: interval,
			Factor:   1.0,
			Jitter:   0.1,
		},
		func(err error) bool {
			log.Logger.Error(err)
			return true
		}, // retry on all failures
		func() error {
			return checkObjectExistsAndConditionTrue(t, restConfig)
		},
	)
	return err
}

// checkObjectExistsAndConditionTrue determines if the object exists
// if so, it further checks if the requested condition is "True"
// based on https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go
func checkObjectExistsAndConditionTrue(t *readinessTask, restCfg *rest.Config) error {
	log.Logger.Trace("looking for object ", t.With.Kind, "/", t.With.Name, " in namespace ", *t.With.Namespace)

	// get object from cluster
	obj, err := getObject(
		&corev1.ObjectReference{
			Kind:      t.With.Kind,
			Name:      t.With.Name,
			Namespace: *t.With.Namespace,
		},
		restCfg,
	)
	if err != nil {
		return err
	}

	// if no condition to check was specified, we can return now
	if t.With.Condition == nil {
		return nil
	}

	// otherwise, find the condition and check that it is "True"
	log.Logger.Trace("looking for condition: ", *t.With.Condition)

	cs, err := getConditionStatus(obj, *t.With.Condition)
	if err != nil {
		return err
	}
	if strings.EqualFold(*cs, string(corev1.ConditionTrue)) {
		return nil
	}
	return errors.New("condition status not True")
}

// getObject finds the object referenced by objRef using the client config restConfig
// uses the dynamic client; ie, retuns an unstructured object
func getObject(objRef *corev1.ObjectReference, restConfig *rest.Config) (*unstructured.Unstructured, error) {
	// dr, err := getDynamicResourceInterface(restConfig, objRef, objRef.Namespace)
	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(restConfig)
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

// getCondition looks for a condition with type conditionType
// This works for objects that follow the recommendation
func getConditionStatus(obj *unstructured.Unstructured, conditionType string) (*string, error) {
	if obj == nil {
		return nil, errors.New("no object")
	}

	resultJson, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	resultObj := make(map[string]interface{})
	err = json.Unmarshal(resultJson, &resultObj)
	if err != nil {
		return nil, err
	}

	// get object status
	objStatusInterface, ok := resultObj["status"]
	if !ok {
		return nil, errors.New("object does not contain a status")
	}
	objStatus := objStatusInterface.(map[string]interface{})

	conditionsInterface, ok := objStatus["conditions"]
	if !ok {
		return nil, errors.New("object status does not contain coditions")
	}
	conditions := conditionsInterface.([]interface{})
	for _, conditionInterface := range conditions {
		condition := conditionInterface.(map[string]interface{})
		cTypeInterface, ok := condition["type"]
		if !ok {
			return nil, errors.New("condition does not have a type")
		}
		cType := cTypeInterface.(string)
		if strings.EqualFold(cType, conditionType) {
			conditionStatusInterface, ok := condition["status"]
			if !ok {
				return nil, fmt.Errorf("condition %s does not have a value", cType)
			}
			conditionStatus := conditionStatusInterface.(string)
			return StringPointer(conditionStatus), nil
		}
	}
	return nil, errors.New("expected condition not found")
}
