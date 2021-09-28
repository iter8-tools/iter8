package readiness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
)

const (
	// TaskName is the name of the readiness task
	TaskName string = "common/readiness"

	// regex for resource names
	dnsLabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

	// maximum length of names
	dnsLabelMaxLength int = 63

	// readiness task default values for params
	defaultInitialDelaySeconds = 5
	defaultNumRetries          = 12
	defaultIntervalSeconds     = 5
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

// regex object
var dnsLabelRegexp = regexp.MustCompile("^" + dnsLabelFmt + "$")

// IsDNSLabel tests for a string that conforms to the definition of a label in
// DNS (RFC 1035/1123).
// The following function is implemented in (very) old versions of k8s util package.
// This function doesn't seem to exist in newer versions, so reimplemented here.
func IsDNSLabel(value string) bool {
	return len(value) <= dnsLabelMaxLength && dnsLabelRegexp.MatchString(value)
}

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
	// InitialDelaySeconds is optional and defaulted to 5 secs. The first check will be performed after this delay.
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
	// NumRetries is optional and defaulted to 12. This is the number of retries that will be attempted after the first check. Total number of trials = 1 + NumRetries.
	NumRetries *int32 `json:"numRetries,omitempty" yaml:"numRetries,omitempty"`
	// IntervalSeconds is optional and defaulted to 5 secs
	// Retries will be attempted periodically every IntervalSeconds
	IntervalSeconds *int32 `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`
	// ObjRefs is a list of K8s objects along with optional readiness conditions
	ObjRefs []ObjRef `json:"objRefs,omitempty" yaml:"objRefs,omitempty"`
}

// ReadinessTask checks existence and readiness of specified resources
type ReadinessTask struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          ReadinessInputs `json:"with" yaml:"with"`
}

// Make creates a readiness task with correct defaults.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, fmt.Errorf("library and task need to be '%s'", TaskName)
	}
	var jsonBytes []byte
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to ReadinessTask
	task := &ReadinessTask{}
	err = json.Unmarshal(jsonBytes, &task)
	if err != nil {
		return nil, err
	}
	// set defaults
	if task.With.InitialDelaySeconds == nil {
		task.With.InitialDelaySeconds = core.Int32Pointer(defaultInitialDelaySeconds)
	}
	if task.With.NumRetries == nil {
		task.With.NumRetries = core.Int32Pointer(defaultNumRetries)
	}
	if task.With.IntervalSeconds == nil {
		task.With.IntervalSeconds = core.Int32Pointer(defaultIntervalSeconds)
	}

	// validate
	for _, o := range task.With.ObjRefs {
		if !IsDNSLabel(o.Name) {
			err = errors.New("object name is malformatted; needs to be a valid DNS label")
			break
		}
	}

	return task, err
}

// command interface is useful for mocking shell commands
type command interface {
	Run() error
	String() string
}

// getCommand returns an instance of the command interface
var getCommand = func(name string, arg ...string) command {
	return exec.Command(name, arg...)
}

// Run checks existence and readiness of K8s objects.
func (t *ReadinessTask) Run(ctx context.Context) error {
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("experiment", exp)

	// add versioninfo objects to task
	// for baseline
	if exp.Spec.VersionInfo != nil {
		if exp.Spec.VersionInfo.Baseline.WeightObjRef != nil {
			objRef := ObjRef{
				Kind:      exp.Spec.VersionInfo.Baseline.WeightObjRef.Kind,
				Namespace: &exp.Spec.VersionInfo.Baseline.WeightObjRef.Namespace,
				Name:      exp.Spec.VersionInfo.Baseline.WeightObjRef.Name,
			}
			t.With.ObjRefs = append(t.With.ObjRefs, objRef)
		}

		// for each candidate
		for _, c := range exp.Spec.VersionInfo.Candidates {
			if c.WeightObjRef != nil {
				objRef := ObjRef{
					Kind:      c.WeightObjRef.Kind,
					Namespace: &c.WeightObjRef.Namespace,
					Name:      c.WeightObjRef.Name,
				}
				t.With.ObjRefs = append(t.With.ObjRefs, objRef)
			}
		}
	}

	log.Info("The task...")
	log.Info(t)
	time.Sleep(time.Duration(*t.With.InitialDelaySeconds) * time.Second)
	// invariant: objIndex is the number of objects that have been checked and found to be good
	objIndex := 0
	for i := 0; i <= int(*t.With.NumRetries); i++ {
		// this inner loop has no busy waiting (sleeps)
		// it will keep going through the objects as much as possible
		for err == nil && objIndex < len(t.With.ObjRefs) {
			// fix namespace
			var namespace string
			if t.With.ObjRefs[i].Namespace == nil {
				namespace = exp.Namespace
			} else {
				namespace = *t.With.ObjRefs[i].Namespace
			}
			// check existence
			script := fmt.Sprintf("kubectl get %s %s -n %s", t.With.ObjRefs[i].Kind, t.With.ObjRefs[i].Name, namespace)
			cmd := getCommand("/bin/bash", "-c", script)

			_, ok := cmd.(*exec.Cmd)
			if ok {
				cmd.(*exec.Cmd).Stdout = os.Stdout
				cmd.(*exec.Cmd).Stderr = os.Stderr
			}

			log.Info("Executing command: " + cmd.String())
			err = cmd.Run()
			if err == nil {
				// check readiness condition if any
				if t.With.ObjRefs[i].WaitFor != nil {
					script := fmt.Sprintf("kubectl wait %s/%s -n %s --for=%s --timeout=0s", t.With.ObjRefs[i].Kind, t.With.ObjRefs[i].Name, namespace, *t.With.ObjRefs[i].WaitFor)
					cmd := getCommand("/bin/bash", "-c", script)

					_, ok := cmd.(*exec.Cmd)
					if ok {
						cmd.(*exec.Cmd).Stdout = os.Stdout
						cmd.(*exec.Cmd).Stderr = os.Stderr
					}

					log.Info("Executing command: " + cmd.String())
					err = cmd.Run()
				}
			}

			if err == nil {
				// advance objIndex
				objIndex++
			}
		}

		if i == int(*t.With.NumRetries) || objIndex == len(t.With.ObjRefs) { // we are done
			break // out of the for loop
		} else {
			// try again later
			time.Sleep(time.Duration(*t.With.IntervalSeconds) * time.Second)
		}
	}

	return err
}
