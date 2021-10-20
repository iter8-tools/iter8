package core

import "time"

type Task interface {
	Run(i int) error
}

// Experiment specification and result
type Experiment struct {
	ExperimentContext
	Tasks  []Task
	Spec   *ExperimentSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Result *ExperimentResult `json:"result,omitempty" yaml:"result,omitempty"`
}

// ExperimentSpec specifies the experiment
type ExperimentSpec struct {
	// Iter8Version is the version of Iter8 used for this experiment spec
	Iter8Version string `json:"iter8Version" yaml:"iter8Version"`

	// Versions are the names of app versions that are assessed in this experiment
	Versions []string `json:"versions" yaml:"versions"`

	// Tasks is the sequence of tasks that constitute this experiment
	Tasks []TaskSpec `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

// TaskSpec contains the specification of a task.
type TaskSpec struct {
	// Task uniquely identifies the task to be executed.
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	// Run is a special type of task meant to run a bash script.
	// TaskSpec must include exactly one of the two fields, run or task.
	Run *string `json:"run,omitempty" yaml:"run,omitempty"`
	// If specifies if this task should be executed.
	// Task will be evaluated if condition specified by if evaluates to true, and not otherwise.
	If *string `json:"if,omitempty" yaml:"if,omitempty"`
	// With holds inputs to this task.
	With map[string]interface{} `json:"with,omitempty" yaml:"with,omitempty"`
}

// ExperimentResult defines the current results from the experiment
type ExperimentResult struct {
	// StartTime is the time when the experiment result is created
	StartTime *time.Time `json:"startTime,omitempty" yaml:"startTime,omitempty"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int32 `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// Failure is true if the experiment failed to complete all the tasks successfully
	Failure bool `json:"failure" yaml:"failure"`

	// Analysis is the latest analysis
	Analysis *Analysis `json:"analysis,omitempty" yaml:"analysis,omitempty"`
}

// Analysis is data from an analytics provider
type Analysis struct {
	// Metrics
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// if not empty, the key in the map is a metric name and the value is a list of observed metric values
	Metrics []map[string][]float64 `json:"metrics,omitempty" yaml:"metrics,omitempty"`

	// Objectives
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// if not empty, the length of an inner slice must match the number of objectives in the assess-versions task
	Objectives [][]bool `json:"objectives,omitempty" yaml:"objectives,omitempty"`

	// Winner is the winning version of the app
	Winner *string `json:"winner,omitempty" yaml:"winner,omitempty"`

	// Weights is the most recently recommended traffic weights
	// if not empty, the length of the slice must match the length of Spec.Versions
	Weights []int32 `json:"weights,omitempty" yaml:"weights,omitempty"`
}

func (e *Experiment) Build() error {
	var err error
	e.Spec, err = e.ExperimentContext.ReadSpec()
	return err
}
