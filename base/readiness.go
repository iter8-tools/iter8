package base

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	log "github.com/iter8-tools/iter8/base/log"
)

//TBD are RFC 1123 names sufficient?

const (
	// task name
	ReadinessTaskName = "k8s-objects-ready"

	// regex for resource names (RFC 1123 label names)
	// see: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	labelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

	// maximum length of names
	// see: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	maxLabelLength int = 63

	// readiness task default values for params
	defaultDelaySeconds         = 5
	defaultRetries              = 12
	defaultRetryIntervalSeconds = 5
)

// regex object
var labelRegexp = regexp.MustCompile("^" + labelFmt + "$")

// ObjRef contains details about a specific K8s object whose existence and readiness will be checked
type ObjRef struct {
	// Kind of the object. Specified in the TYPE[.VERSION][.GROUP] format used by `kubectl`
	// See https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#get
	Kind string `json:"kind" yaml:"kind"`
	// Namespace of the object. Optional. If left unspecified, this will be defaulted to the namespace of the experiment
	Namespace *string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Name of the object
	Name string `json:"name" yaml:"name"`
	// Wait for condition. Optional.
	// Any value that is accepted by the --for flag of the `kubectl wait` command can be specified.
	// See https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#wait
	WaitFor *string `json:"waitFor,omitempty" yaml:"waitFor,omitempty"`
}

// ReadinessInputs contains a list of K8s object references along with
// optional readiness conditions for them. The inputs also specify the delays
// and retries involved in the existence and readiness checks.
// This task will also check for existence of objects specified
// in the VersionInfo field of the experiment.
type ReadinessInputs struct {
	// DealySeconds is the time in seconds before the first check is made.
	// Optional; default is 5 seconds.
	DelaySeconds *int `json:"delay,omitempty" yaml:"delay,omitempty"`
	// Retries the number of attempts to check waitfor condition.
	// Optional; default is 12.
	Retries *int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// RetryIntervalSeconds time in seconds between retries.
	// Optional; default is 5 seconds.
	RetryIntervalSeconds *int `json:"retryInterval,omitempty" yaml:"retryInterval,omitempty"`
	// ObjRefs is a list of Kubernetes objects and an optional readiness conditions
	Objects []ObjRef `json:"objects,omitempty" yaml:"objects,omitempty"`
}

// ReadinessTask checks existence and readiness of specified resources
type ReadinessTask struct {
	TaskMeta
	With ReadinessInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the readiness task
func (t *ReadinessTask) initializeDefaults() {
	log.Logger.Info("initailizing defaults")
	// get value from With or default values
	if t.With.DelaySeconds == nil {
		t.With.DelaySeconds = intPointer(defaultDelaySeconds)
	}
	if t.With.Retries == nil {
		t.With.Retries = intPointer(defaultRetries)
	}
	if t.With.RetryIntervalSeconds == nil {
		t.With.RetryIntervalSeconds = intPointer(defaultRetryIntervalSeconds)
	}
}

// validate task inputs
func (t *ReadinessTask) validateInputs() error {
	log.Logger.Info("validating inputs")
	for _, obj := range t.With.Objects {
		if len(obj.Name) > maxLabelLength || !labelRegexp.MatchString(obj.Name) {
			return errors.New("invalid object name; It must be a valid DNS label")
		}
	}

	return nil
}

// execute the task
func (t *ReadinessTask) Run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	time.Sleep(time.Duration(*t.With.DelaySeconds) * time.Second)

	successfullyVerified := 0
	for attempt := 0; attempt < *t.With.Retries; attempt++ {
		successfullyVerified = 0
		for _, obj := range t.With.Objects {
			var ns string
			if obj.Namespace == nil {
				ns = "default"
				// TODO use context to determine this
			} else {
				ns = *obj.Namespace
			}
			script := fmt.Sprintf("kubectl get %s/%s -n %s", obj.Kind, obj.Name, ns)
			cmd := exec.Command("/bin/bash", "-c", script)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Logger.Info("Executing command: " + cmd.String())
			err := cmd.Run()
			if err != nil {
				continue
			}
			if obj.WaitFor != nil {
				script := fmt.Sprintf("kubectl wait %s/%s -n %s --for=%s --timeout=0s", obj.Kind, obj.Name, ns, *obj.WaitFor)
				cmd := exec.Command("/bin/bash", "-c", script)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				log.Logger.Info("Executing command: " + cmd.String())
				err = cmd.Run()

			}
			if err == nil {
				successfullyVerified++
			}
		} // for _, obj := range

		// if found all objecgts OR completed last attempt (in this case skip the sleep)
		if successfullyVerified == len(t.With.Objects) || attempt == *t.With.Retries {
			break
		}

		time.Sleep(time.Duration(*t.With.RetryIntervalSeconds) * time.Second)
	} // for attempt

	if successfullyVerified != len(t.With.Objects) {
		return errors.New("some objects not ready")
		// TODO would it be helpful to know which objects?
	}
	return nil
}
