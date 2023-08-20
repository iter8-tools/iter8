package base

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/antonmedv/expr"
	log "github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/time"
)

// Task is the building block of an experiment spec
// An experiment spec is a sequence of tasks
type Task interface {
	// validateInputs for this task
	validateInputs() error

	// initializeDefaults of the input values to this task
	initializeDefaults()

	// run this task
	run(exp *Experiment) error
}

// ExperimentSpec specifies the set of tasks in this experiment
type ExperimentSpec []Task

// ExperimentMetadata species the name and namespace of the experiment
// Used in http and grpc tasks to send the name and namespace to the metrics server
type ExperimentMetadata struct {
	// Name is the name of the experiment
	Name string `json:"name" yaml:"name"`

	// Namespace is the namespace the experiment was deployed in
	Namespace string `json:"namespace" yaml:"namespace"`
}

// Experiment struct containing spec and result
type Experiment struct {
	Metadata ExperimentMetadata `json:"metadata" yaml:"metadata"`

	// Spec is the sequence of tasks that constitute this experiment
	Spec ExperimentSpec `json:"spec" yaml:"spec"`

	// Result is the current results from this experiment.
	// The experiment may not have completed in which case results may be partial.
	Result *ExperimentResult `json:"result" yaml:"result"`

	// driver enables interacting with experiment result stored externally
	driver Driver
}

// ExperimentResult defines the current results from the experiment
type ExperimentResult struct {
	// Name is the name of this experiment
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Namespace is the namespace of this experiment
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Revision of this experiment
	Revision int `json:"revision,omitempty" yaml:"revision,omitempty"`

	// StartTime is the time when the experiment run started
	StartTime time.Time `json:"startTime" yaml:"startTime"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// Failure is true if any of its tasks failed
	Failure bool `json:"failure" yaml:"failure"`

	// Insights produced in this experiment
	Insights *Insights `json:"insights,omitempty" yaml:"insights,omitempty"`

	// Iter8Version is the version of Iter8 CLI that created this result object
	Iter8Version string `json:"iter8Version" yaml:"iter8Version"`
}

// Insights records the number of versions in this experiment
type Insights struct {
	// NumVersions is the number of app versions detected by Iter8
	NumVersions int `json:"numVersions" yaml:"numVersions"`

	// VersionNames is list of version identifiers if known
	VersionNames []VersionInfo `json:"versionNames" yaml:"versionNames"`

	// TaskData is a map of task names to the data produced by said task
	TaskData map[string]interface{} `json:"taskData" yaml:"taskData"`
}

// VersionInfo is basic information about a version
type VersionInfo struct {
	// Version name
	Version string `json:"version" yaml:"version"`

	// Track identifier assigned to version
	Track string `json:"track" yaml:"track"`
}

// TaskMeta provides common fields used across all tasks
type TaskMeta struct {
	// Task is the name of the task
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	// Run is the script used in a run task
	// Specify either Task or Run but not both
	Run *string `json:"run,omitempty" yaml:"run,omitempty"`
	// If is the condition used to determine if this task needs to run
	// If the condition is not satisfied, then it is skipped in an experiment
	If *string `json:"if,omitempty" yaml:"if,omitempty"`
}

// taskMetaWith enables unmarshaling of tasks
type taskMetaWith struct {
	// TaskMeta has fields common to all tasks
	TaskMeta
	// With is the raw representation of task inputs
	With map[string]interface{} `json:"with,omitempty" yaml:"with,omitempty"`
}

// UnmarshalJSON will unmarshal an experiment spec from bytes
// This is a custom JSON unmarshaler
func (s *ExperimentSpec) UnmarshalJSON(data []byte) error {
	var v []taskMetaWith
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	log.Logger.Tracef("unmarshaled %v tasks into task meta", len(v))

	for _, t := range v {
		if (t.Task == nil || len(*t.Task) == 0) && (t.Run == nil) {
			err := fmt.Errorf("invalid task found without a task name or a run command")
			log.Logger.Error(err)
			return err
		}

		// get byte data for this task
		tBytes, _ := json.Marshal(t)
		var tsk Task
		// this is a run task
		if t.Run != nil {
			log.Logger.Debug("found run task: ", *t.Run)
			rt := &runTask{}
			if err := json.Unmarshal(tBytes, rt); err != nil {
				e := errors.New("json unmarshal error")
				log.Logger.WithStackTrace(err.Error()).Error(e)
				return e
			}
			tsk = rt
		} else {
			// this is some other task
			switch *t.Task {
			case ReadinessTaskName:
				rt := &readinessTask{}
				if err := json.Unmarshal(tBytes, rt); err != nil {
					e := errors.New("json unmarshal error")
					log.Logger.WithStackTrace(err.Error()).Error(e)
					return e
				}
				tsk = rt
			case CollectHTTPTaskName:
				cht := &collectHTTPTask{}
				if err := json.Unmarshal(tBytes, cht); err != nil {
					e := errors.New("json unmarshal error")
					log.Logger.WithStackTrace(err.Error()).Error(e)
					return e
				}
				tsk = cht
			case CollectGRPCTaskName:
				cgt := &collectGRPCTask{}
				if err := json.Unmarshal(tBytes, cgt); err != nil {
					e := errors.New("json unmarshal error")
					log.Logger.WithStackTrace(err.Error()).Error(e)
					return e
				}
				tsk = cgt
			case NotifyTaskName:
				nt := &notifyTask{}
				if err := json.Unmarshal(tBytes, nt); err != nil {
					e := errors.New("json unmarshal error")
					log.Logger.WithStackTrace(err.Error()).Error(e)
					return e
				}
				tsk = nt
			default:
				log.Logger.Error("unknown task: " + *t.Task)
				return errors.New("unknown task: " + *t.Task)
			}
		}
		n := append(*s, tsk)
		*s = n
		log.Logger.Trace("appended to experiment spec")
	}
	log.Logger.Trace("constructed experiment spec of length: ", len(*s))
	return nil
}

// TrackVersionStr creates a string of version name/track for display purposes
func (in *Insights) TrackVersionStr(i int) string {
	// if VersionNames not defined or all fields empty return default "version i"
	if in.VersionNames == nil ||
		len(in.VersionNames) == 0 ||
		len(in.VersionNames[i].Version)+len(in.VersionNames[i].Track) == 0 {
		return fmt.Sprintf("version %d", i)
	}

	if len(in.VersionNames[i].Track) == 0 {
		// version not ""
		return in.VersionNames[i].Version
	}

	if len(in.VersionNames[i].Version) == 0 {
		// track not ""
		return in.VersionNames[i].Track
	}

	return in.VersionNames[i].Track + " (" + in.VersionNames[i].Version + ")"
}

// initResults initializes the results section of an experiment
func (exp *Experiment) initResults(revision int) {
	exp.Result = &ExperimentResult{
		Name:              exp.Metadata.Name,
		Namespace:         exp.Metadata.Namespace,
		Revision:          revision,
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
		Iter8Version:      MajorMinor,
	}
}

// initInsightsWithNumVersions is also going to initialize insights data structure
// insights data structure contains metrics data structures, so this will also
// init metrics
func (r *ExperimentResult) initInsightsWithNumVersions(n int) error {
	if r.Insights == nil {
		r.Insights = &Insights{
			NumVersions: n,
			TaskData:    map[string]interface{}{},
		}
	} else {
		if r.Insights.NumVersions != n {
			err := fmt.Errorf("inconsistent number for app versions; old (%v); new (%v)", r.Insights.NumVersions, n)
			log.Logger.Error(err)
			return err
		}
	}

	return nil
}

// Driver enables interacting with experiment result stored externally
type Driver interface {
	// Read the experiment
	Read() (*Experiment, error)

	// deprecated
	// Write the experiment
	Write(e *Experiment) error

	// GetRevision returns the experiment revision
	GetRevision() int
}

// Completed returns true if the experiment is complete
func (exp *Experiment) Completed() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.NumCompletedTasks == len(exp.Spec) {
				return true
			}
		}
	}
	return false
}

// NoFailure returns true if no task in the experiment has failed
func (exp *Experiment) NoFailure() bool {
	return exp != nil && exp.Result != nil && !exp.Result.Failure
}

// run the experiment
func (exp *Experiment) run(driver Driver) error {
	var err error

	exp.driver = driver
	if exp.Result == nil {
		err = errors.New("experiment with nil result section cannot be run")
		log.Logger.Error(err)
		return err
	}
	log.Logger.Debug("exp result exists now ... ")

	log.Logger.Debugf("attempting to execute %v tasks", len(exp.Spec))
	for i, t := range exp.Spec {
		log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + ": started")

		shouldRun := true
		// if task has a condition
		if cond := getIf(t); cond != nil {
			// condition evaluates to false ... then shouldRun is false
			program, err := expr.Compile(*cond, expr.Env(exp), expr.AsBool())
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to compile if clause")
				return err
			}

			output, err := expr.Run(program, exp)
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("unable to run if clause")
				return err
			}

			shouldRun = output.(bool)
		}
		if shouldRun {
			err = t.run(exp)
			if err != nil {
				log.Logger.Error("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + ": " + "failure")
				exp.failExperiment()

				err = driver.Write(exp)
				if err != nil {
					return err
				}
			}
			log.Logger.Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + ": " + "completed")
		} else {
			log.Logger.WithStackTrace(fmt.Sprint("false condition: ", *getIf(t))).Info("task " + fmt.Sprintf("%v: %v", i+1, *getName(t)) + ": " + "skipped")
		}

		exp.incrementNumCompletedTasks()

		err = driver.Write(exp)
		if err != nil {
			return err
		}
	}
	return nil
}

// failExperiment sets the experiment failure status to true
func (exp *Experiment) failExperiment() {
	exp.Result.Failure = true
}

// incrementNumCompletedTasks increments the number of completed tasks in the experiment
func (exp *Experiment) incrementNumCompletedTasks() {
	exp.Result.NumCompletedTasks++
}

// getIf returns the condition (if any) which determine
// whether of not if this task needs to run
func getIf(t Task) *string {
	var jsonBytes []byte
	var tm TaskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to TaskMeta
	_ = json.Unmarshal(jsonBytes, &tm)
	return tm.If
}

// getName returns the name of this task
func getName(t Task) *string {
	var jsonBytes []byte
	var tm TaskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to TaskMeta
	_ = json.Unmarshal(jsonBytes, &tm)

	if tm.Task == nil {
		if tm.Run != nil {
			return StringPointer(RunTaskName)
		}
	} else {
		return tm.Task
	}
	log.Logger.Error("task spec with no name or run value")
	return nil
}

// BuildExperiment builds an experiment
func BuildExperiment(driver Driver) (*Experiment, error) {
	e, err := driver.Read()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// RunExperiment runs an experiment
func RunExperiment(driver Driver) error {
	var exp *Experiment
	var err error
	if exp, err = BuildExperiment(driver); err != nil {
		return err
	}

	result, _ := json.Marshal(exp.Result)
	log.Logger.Trace("Initializing result", string(result))

	exp.initResults(driver.GetRevision())

	result, _ = json.Marshal(exp.Result)
	log.Logger.Trace("initialized result", string(result))

	return exp.run(driver)
}
