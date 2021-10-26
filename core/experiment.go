package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"fortio.org/fortio/fhttp"
	"github.com/ghodss/yaml"
)

// Experiment specification and result
type Experiment struct {
	TaskMaker `json:"-" yaml:"-"`
	Name      string            `json:"name" yaml:"name"`
	Spec      *ExperimentSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Result    *ExperimentResult `json:"result,omitempty" yaml:"result,omitempty"`
}

// ExperimentSpec specifies the experiment
type ExperimentSpec struct {
	// Iter8Version is the version of Iter8 used for this experiment spec
	Iter8Version string `json:"iter8Version" yaml:"iter8Version"`

	// Versions are the names of app versions that are assessed in this experiment
	Versions []string `json:"versions" yaml:"versions"`

	// Tasks is the sequence of tasks that constitute this experiment
	Tasks []TaskSpec `json:"tasks,omitempty" yaml:"tasks,omitempty"`

	// tasks is the runnable representation of tasks
	tasks []Task
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

	// DefaultFilePath is the default path to experiment file
	DefaultFilePath = "experiment.yaml"
)

var (
	// Path to experiment file
	// this variable is not intended to be modified in tests, and nowhere else
	filePath = "experiment.yaml"
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
	TestingPattern *TestingPatternType `json:"testingPattern,omitempty" yaml:"testingPattern,omitempty"`

	// FortioMetrics populated by the collect-fortio-metrics task
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	FortioMetrics []*fhttp.HTTPRunnerResults `json:"fortioMetrics,omitempty" yaml:"fortioMetrics,omitempty"`

	// Metrics
	// if not empty, the length of the outer slice must match the length of Spec.Versions
	// each key in the map is a metric name
	// values are all the observed values of a metric until this point
	Metrics []map[string][]float64 `json:"metrics,omitempty" yaml:"metrics,omitempty"`

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

// Build an experiment from file
func (e *Experiment) Build(withResult bool) error {
	// read it in
	Logger.Trace("build called")
	newExp, err := Read()
	if err != nil {
		return err
	}
	e.Name, e.Spec, e.Result = newExp.Name, newExp.Spec, newExp.Result
	if !withResult {
		e.Result = &ExperimentResult{}
	}
	// make tasks
	for i, ts := range e.Spec.Tasks {
		Logger.Trace(fmt.Sprintf("unmarshaling task %v", i))
		t, err := e.TaskMaker.Make(&ts)
		if err != nil {
			return err
		}
		e.Spec.tasks = append(e.Spec.tasks, t)
	}

	return err
}

func (e *Experiment) setStartTime() error {
	if e.Result == nil {
		Logger.Warn("setStartTime called on an experiment object without results")
		e.initResults()
	}
	e.Result.StartTime = TimePointer(time.Now())
	return Write(e)
}

func (e *Experiment) failExperiment() error {
	if e.Result == nil {
		Logger.Warn("failExperiment called on an experiment object without results")
		e.initResults()
	}
	e.Result.Failure = true
	return Write(e)
}

func (e *Experiment) incrementNumCompletedTasks() error {
	if e.Result == nil {
		Logger.Warn("incrementNumCompletedTasks called on an experiment object without results")
		e.initResults()
	}
	e.Result.NumCompletedTasks++
	return Write(e)
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
	return Write(e)
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
	return Write(e)
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
	return Write(e)
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
	return Write(e)
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
	return Write(e)
}

// Read an experiment from a file
func Read() (*Experiment, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return nil, errors.New("unable to read experiment file")
	}
	e := Experiment{}
	err = yaml.Unmarshal(yamlFile, &e)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to unmarshal experiment")
		return nil, err
	}
	return &e, err
}

// Write an experiment to a file
func Write(r *Experiment) error {
	rBytes, err := yaml.Marshal(r)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to marshal experiment")
		return errors.New("unable to marshal experiment")
	}
	err = ioutil.WriteFile(filePath, rBytes, 0664)
	if err != nil {
		Logger.WithStackTrace(err.Error()).Error("unable to write experiment file")
		return err
	}
	return err
}
