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

//TBD Are RFC 1123 names sufficient?

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

// ReadinessInputs contains a list of K8s object references along with
// optional readiness conditions for them. The inputs also specify the delays
// and retries involved in the existence and readiness checks.
// This task will also check for existence of objects specified
// in the VersionInfo field of the experiment.
type ReadinessInputs struct {
	// DealySeconds is the time in seconds before the first check is made.
	// Optional; default is 5 seconds.
	Delay *int `json:"delay,omitempty" yaml:"delay,omitempty"`
	// Retries the number of attempts to check waitfor condition.
	// Optional; default is 12.
	Retries *int `json:"retries,omitempty" yaml:"retries,omitempty"`
	// RetryIntervalSeconds time in seconds between retries.
	// Optional; default is 5 seconds.
	RetryInterval *int `json:"retryInterval,omitempty" yaml:"retryInterval,omitempty"`
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

// ReadinessTask checks existence and readiness of specified resources
type ReadinessTask struct {
	TaskMeta
	With ReadinessInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the readiness task
func (t *ReadinessTask) initializeDefaults() {
	// get value from With or default values
	if t.With.Delay == nil {
		t.With.Delay = intPointer(defaultDelaySeconds)
	}
	if t.With.Retries == nil {
		t.With.Retries = intPointer(defaultRetries)
	}
	if t.With.RetryInterval == nil {
		t.With.RetryInterval = intPointer(defaultRetryIntervalSeconds)
	}
}

// validate task inputs
func (t *ReadinessTask) validateInputs() error {
	name := t.With.Name
	if len(name) > maxLabelLength || !labelRegexp.MatchString(name) {
		return errors.New("invalid object name; It must be a valid DNS label")
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

	time.Sleep(time.Duration(*t.With.Delay) * time.Second)

	// TODO use context to determine this
	ns := "default"
	if t.With.Namespace != nil {
		ns = *t.With.Namespace
	}

	successfullyVerified := false
	for attempt := 0; attempt < *t.With.Retries; attempt++ {
		getCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("kubectl get %s/%s -n %s", t.With.Kind, t.With.Name, ns))
		getCmd.Stdout = os.Stdout
		getCmd.Stderr = os.Stderr
		log.Logger.Info("Executing command: " + getCmd.String())
		err := getCmd.Run()
		if err == nil {
			if t.With.WaitFor != nil {
				waitCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("kubectl wait %s/%s -n %s --for=%s --timeout=0s", t.With.Kind, t.With.Name, ns, *t.With.WaitFor))
				waitCmd.Stdout = os.Stdout
				waitCmd.Stderr = os.Stderr
				log.Logger.Info("Executing command: " + waitCmd.String())
				err = waitCmd.Run()
			}
		}
		if err == nil {
			successfullyVerified = true
			break
		}

		// if not verified and not final attempt; sleep for retryInterval
		if !successfullyVerified && attempt != *t.With.Retries {
			time.Sleep(time.Duration(*t.With.RetryInterval) * time.Second)
		}
	} // for attempt

	if !successfullyVerified {
		return errors.New("object not ready")
	}
	return nil
}
