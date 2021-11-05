package base

import (
	"errors"
	"fmt"
	"time"

	"fortio.org/fortio/fhttp"
	log "github.com/iter8-tools/iter8/base/log"
)

// Experiment specification and result
type Experiment struct {
	// Tasks is the sequence of tasks that constitute this experiment
	Tasks  []TaskSpec        `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Result *ExperimentResult `json:"result,omitempty" yaml:"result,omitempty"`
}

// Task objects can be run
type Task interface {
	Run(exp *Experiment) error
}

// ExperimentResult defines the current results from the experiment
type ExperimentResult struct {
	// StartTime is the time when the experiment result is created
	StartTime *time.Time `json:"startTime,omitempty" yaml:"startTime,omitempty"`

	// NumAppVersions is the number of app versions detected by Iter8 in this experiment
	NumAppVersions *int `json:"numAppVersions,omitempty" yaml:"numVersions,omitempty"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

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

	// Valid is the set of all versions that satisfy objectives
	Valid []string `json:"valid,omitempty" yaml:"valid,omitempty"`

	// Winner is the winning version of the app
	Winner *string `json:"winner,omitempty" yaml:"winner,omitempty"`
}

type taskMeta struct {
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	Run  *string `json:"run,omitempty" yaml:"run,omitempty"`
	If   *string `json:"if,omitempty" yaml:"if,omitempty"`
}

// TaskSpec has information needed to construct a Task
type TaskSpec struct {
	taskMeta
	With map[string]interface{} `json:"with,omitempty" yaml:"with,omitempty"`
}

// func (t *taskMeta) bytes() []byte {
// 	b, _ := json.Marshal(t)
// 	return b
// }

// String converts the experiment into a yaml string
// func (e *Experiment) String() string {
// 	out, _ := yaml.Marshal(e)
// 	return string(out)
// }

// setTestingPattern sets the testing pattern in the experiment results
func (e *Experiment) setTestingPattern(c *Criteria) error {
	if e.Result == nil {
		log.Logger.Warn("setTestingPattern called on an experiment object without results")
		e.InitResults()
	}
	if e.Result.Analysis == nil {
		log.Logger.Warn("setTestingPattern called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	if c == nil || c.Objectives == nil || len(c.Objectives) == 0 {
		e.Result.Analysis.TestingPattern = testingPatternPointer(TestingPatternNone)
	} else {
		e.Result.Analysis.TestingPattern = testingPatternPointer(TestingPatternSLOValidation)
	}
	return nil
}

// setObjectives sets objective assessment portion of the analysis
func (e *Experiment) setObjectives(objs [][]bool) error {
	if e.Result == nil {
		log.Logger.Warn("setObjectives called on an experiment object without results")
		e.InitResults()
	}
	if e.Result.Analysis == nil {
		log.Logger.Warn("setObjectives called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.Objectives = objs
	return nil
}

// setWinner sets the winning version
func (e *Experiment) setWinner(winner *string) error {
	if e.Result == nil {
		log.Logger.Warn("setWinner called on an experiment object without results")
		e.InitResults()
	}
	if e.Result.Analysis == nil {
		log.Logger.Warn("setWinner called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.Winner = winner
	return nil
}

// setValid sets the valid versions
func (e *Experiment) setValid(valid []string) error {
	if e.Result == nil {
		log.Logger.Warn("setValid called on an experiment object without results")
		e.InitResults()
	}
	if e.Result.Analysis == nil {
		log.Logger.Warn("setValid called on an experiment object without analysis")
		e.Result.initAnalysis()
	}
	e.Result.Analysis.Valid = valid
	return nil
}

func (r *ExperimentResult) initAnalysis() {
	r.Analysis = &Analysis{}
}

func (r *ExperimentResult) initNumAppVersions(n int) {
	r.NumAppVersions = intPointer(n)
}

func (e *Experiment) InitResults() {
	e.Result = &ExperimentResult{
		StartTime:         timePointer(time.Now()),
		NumCompletedTasks: 0,
		Failure:           false,
		Analysis:          nil,
	}
	e.Result.initAnalysis()
}

// updateMetricForVersion updates value of a given metric for a given version
func (e *Experiment) updateMetricForVersion(m string, i int, val float64) error {
	if e.Result == nil {
		log.Logger.Error("updateMetricForVersion called on an experiment object without results")
		return errors.New("updateMetricForVersion called on an experiment object without results")
	}
	if e.Result.Analysis == nil {
		log.Logger.Error("updateMetricForVersion called on an experiment object without analysis")
		return errors.New("updateMetricForVersion called on an experiment object without analysis")
	}
	if e.Result.NumAppVersions == nil {
		log.Logger.Error("updateMetricForVersion called on an experiment object without number of app versions uninitialized")
		return errors.New("updateMetricForVersion called on an experiment object without number of app versions uninitialized")
	}
	if e.Result.Analysis.Metrics == nil {
		e.Result.Analysis.Metrics = make([]map[string][]float64, *e.Result.NumAppVersions)
	}
	if i >= *e.Result.NumAppVersions {
		log.Logger.Error("updateMetricForVersion called for version ", i, " but number of app versions is set to ", *e.Result.NumAppVersions)
		return errors.New(fmt.Sprint("updateMetricForVersion called for version ", i, " but number of app versions is set to ", *e.Result.NumAppVersions))
	}
	if e.Result.Analysis.Metrics[i] == nil {
		e.Result.Analysis.Metrics[i] = make(map[string][]float64)
	}
	if _, ok := e.Result.Analysis.Metrics[i][m]; !ok {
		e.Result.Analysis.Metrics[i][m] = []float64{}
	}
	e.Result.Analysis.Metrics[i][m] = append(e.Result.Analysis.Metrics[i][m], val)
	return nil
}
