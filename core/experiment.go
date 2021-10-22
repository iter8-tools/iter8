package core

import (
	"fmt"
	"time"

	"fortio.org/fortio/fhttp"
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

// TestingPatternType identifies the type of experiment
type TestingPatternType string

const (
	// TestingPatternSLOValidation is an SLO validation experiment
	TestingPatternSLOValidation TestingPatternType = "SLOValidation"

	// TestingPatternNone implies no testing of any kind
	TestingPatternNone TestingPatternType = "None"
)

// Criteria is list of criteria to be evaluated throughout the experiment
type Criteria struct {
	// Objectives is a list of conditions on metrics that must be tested on each loop of the experiment.
	// Failure of an objective might reduces the likelihood that a version will be selected as the winning version.
	Objectives []Objective `json:"objectives,omitempty" yaml:"objectives,omitempty"`
}

// Objective is a service level objective
type Objective struct {
	// Metric is the name of the metric resource that defines the metric to be measured.
	// If the value contains a "/", the prefix will be considered to be a namespace name.
	// If the value does not contain a "/", the metric should be defined either in the same namespace
	// or in the default domain namespace (defined as a property of iter8 when installed).
	// The experiment namespace takes precedence.
	Metric string `json:"metric" yaml:"metric"`

	// UpperLimit is the maximum acceptable value of the metric.
	UpperLimit *float64 `json:"upperLimit,omitempty" yaml:"upperLimit,omitempty"`

	// LowerLimit is the minimum acceptable value of the metric.
	LowerLimit *float64 `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty"`
}

// Analysis is data from an analytics provider
type Analysis struct {
	// TestingPattern is the type of this experiment
	TestingPattern *TestingPatternType
	// FortioMetrics populated by the collect-fortio-metrics task
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	FortioMetrics []*fhttp.HTTPRunnerResults

	// Metrics
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// each key in the map is a metric name
	// values are all the observed values of a metric until this point
	Metrics []map[string][]float64

	// Objectives
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// if not empty, the length of an inner slice must match the number of objectives in the assess-versions task
	Objectives [][]bool `json:"objectives,omitempty" yaml:"objectives,omitempty"`

	// Winner is the winning version of the app
	Winner *string `json:"winner,omitempty" yaml:"winner,omitempty"`
}

// String converts the experiment into a yaml string
func (e *Experiment) String() string {
	out, _ := yaml.Marshal(e)
	return string(out)
}

// BuildSpec creates the experiment spec from the spec file
func (e *Experiment) BuildSpec() error {
	Logger.Trace("build spec called")
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
		e.Tasks = append(e.Tasks, t)
	}

	return err
}

// Build creates the experiment from the spec and result files
func (e *Experiment) Build() error {
	err := e.BuildSpec()
	if err != nil {
		return err
	}
	// err = e.BuildResult()
	return err
}

func (e *Experiment) setStartTime() error {
	if e.Result == nil {
		Logger.Warn("setStartTime called on an experiment object without results")
		e.initResults()
	}
	e.Result.StartTime = TimePointer(time.Now())
	return e.ExperimentContext.WriteResult(e.Result)
}

func (e *Experiment) failExperiment() error {
	if e.Result == nil {
		Logger.Warn("failExperiment called on an experiment object without results")
		e.initResults()
	}
	e.Result.Failure = true
	return e.ExperimentContext.WriteResult(e.Result)
}

func (e *Experiment) incrementNumCompletedTasks() error {
	if e.Result == nil {
		Logger.Warn("incrementNumCompletedTasks called on an experiment object without results")
		e.initResults()
	}
	e.Result.NumCompletedTasks++
	return e.ExperimentContext.WriteResult(e.Result)
}

// SetFortioMetrics sets fortio metrics in the experiment results
func (e *Experiment) SetFortioMetrics(fm []*fhttp.HTTPRunnerResults) error {
	if e.Result == nil {
		Logger.Warn("SetFortioMetrics called on an experiment object without results")
		e.initResults()
	}
	if e.Result.Analysis == nil {
		Logger.Warn("SetFortioMetrics called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.FortioMetrics = fm
	return e.ExperimentContext.WriteResult(e.Result)
}

// SetTestingPattern sets the testing pattern in the experiment results
func (e *Experiment) SetTestingPattern(c *Criteria) error {
	if e.Result == nil {
		Logger.Warn("SetTestingPattern called on an experiment object without results")
		e.initResults()
	}
	if e.Result.Analysis == nil {
		Logger.Warn("SetTestingPattern called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	if c == nil || c.Objectives == nil || len(c.Objectives) == 0 {
		e.Result.Analysis.TestingPattern = TestingPatternPointer(TestingPatternNone)
	} else {
		e.Result.Analysis.TestingPattern = TestingPatternPointer(TestingPatternSLOValidation)
	}
	return e.ExperimentContext.WriteResult(e.Result)
}

// SetObjectives sets objective assessment portion of the analysis
func (e *Experiment) SetObjectives(objs [][]bool) error {
	if e.Result == nil {
		Logger.Warn("SetObjectivesSetObjectives called on an experiment object without results")
		e.initResults()
	}
	if e.Result.Analysis == nil {
		Logger.Warn("SetObjectivesSetObjectives called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.Objectives = objs
	return e.ExperimentContext.WriteResult(e.Result)
}

// SetWinner sets the winning version
func (e *Experiment) SetWinner(winner *string) error {
	if e.Result == nil {
		Logger.Warn("SetWinner called on an experiment object without results")
		e.initResults()
	}
	if e.Result.Analysis == nil {
		Logger.Warn("SetWinner called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.Winner = winner
	return e.ExperimentContext.WriteResult(e.Result)
}

func (r *ExperimentResult) initAnalysis() {
	r.Analysis = &Analysis{}
}

func (e *Experiment) initResults() {
	e.Result = &ExperimentResult{
		StartTime:         TimePointer(time.Now()),
		NumCompletedTasks: 0,
		Failure:           false,
		Analysis:          nil,
	}
	e.Result.initAnalysis()
}

// UpdateMetricForVersion updates value of a given metric for a given version
func (e *Experiment) UpdateMetricForVersion(m string, i int, val float64) error {
	if e.Result == nil {
		Logger.Warn("UpdateMetricForVersion called on an experiment object without results")
		e.initResults()
	}
	if e.Result.Analysis == nil {
		Logger.Warn("UpdateMetricForVersion called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	if e.Result.Analysis.Metrics == nil {
		e.Result.Analysis.Metrics = make([]map[string][]float64, len(e.Spec.Versions))
	}
	if e.Result.Analysis.Metrics[i] == nil {
		e.Result.Analysis.Metrics[i] = make(map[string][]float64)
	}
	if _, ok := e.Result.Analysis.Metrics[i][m]; !ok {
		e.Result.Analysis.Metrics[i][m] = []float64{}
	}
	e.Result.Analysis.Metrics[i][m] = append(e.Result.Analysis.Metrics[i][m], val)
	return e.ExperimentContext.WriteResult(e.Result)
}
