package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/iter8-tools/iter8/base/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

const (
	// ReadinessTaskName is the task name
	ReadinessTaskName = "k8s-objects-ready"

	// defaultTimeout is default timeout for readiness command
	defaultTimeout = "10s"
)

// ReadinessInputs identifies the K8s object to test for existence and
// the (optional) condition that should be tested (succeeds if true).
type readinessInputs struct {
	// APIVersion of the object. Optional. If unspecified it will be defaulted to ""
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	// Resource type of the object.
	Resource string `json:"resource" yaml:"resource"`
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
// at present all validation at initialization
func (t *readinessTask) validateInputs() error {
	return nil
}

// run executes the task
func (t *readinessTask) run(exp *Experiment) error {
	// validation
	err := t.validateInputs()
	if err != nil {
		return err
	}

	// initialization
	t.initializeDefaults()
	kd.initKube()
	// set Namespace (from context) if not already set
	if t.With.Namespace == nil {
		t.With.Namespace = kd.Namespace
	}
	timeout, err := time.ParseDuration(*t.With.Timeout)
	if err != nil {
		e := errors.New("invalid format for timeout")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}
	log.Logger.Trace("duration is ", timeout)

	// do the work: check for object and condition
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
			return checkObjectExistsAndConditionTrue(t, kd.RestConfig)
		},
	)
	return err
}

// checkObjectExistsAndConditionTrue determines if the object exists
// if so, it further checks if the requested condition is "True"
func checkObjectExistsAndConditionTrue(t *readinessTask, restCfg *rest.Config) error {
	log.Logger.Trace("looking for ", t.With.Resource, " resource: ", t.With.Name, " in namespace ", *t.With.Namespace)

	obj, err := kd.DynamicClient.Resource(gvr(&t.With)).Namespace(*t.With.Namespace).Get(context.Background(), t.With.Name, metav1.GetOptions{})
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

func gvr(objRef *readinessInputs) schema.GroupVersionResource {
	if gv, err := schema.ParseGroupVersion(objRef.APIVersion); err == nil {
		if gv.Group == "core" {
			gv.Group = ""
		}
		if gv.Version == "" {
			gv.Version = "v1"
		}
		return schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: objRef.Resource}
	}
	return schema.GroupVersionResource{Resource: objRef.Resource}
}

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
		return nil, errors.New("object status does not contain conditions")
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
