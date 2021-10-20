package core

import (
	"fmt"
	"time"

	"github.com/ghodss/yaml"
)

// Experiment specification and result
type Experiment struct {
	ExperimentContext
	TaskMaker
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

func (e *Experiment) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

func (e *Experiment) Build() error {
	Logger.Trace("build called")
	var err error
	e.Spec, err = e.ExperimentContext.ReadSpec()
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to read experiment spec")
		return err
	}
	Logger.WithStackTrace(e.String()).Trace("unmarshaled experiment")
	for i, ts := range e.Spec.Tasks {
		Logger.Trace(fmt.Sprintf("unmarshaling task %v", i))
		t, err := e.TaskMaker.Make(&ts)
		if err != nil {
			Logger.WithStackTrace(err.Error()).Error("unable to unmarshal task")
			return err
		}
		Logger.Info(fmt.Sprintf("task %v", i))
		e.Tasks = append(e.Tasks, t)
	}
	return err
}
