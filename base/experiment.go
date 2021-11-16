package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	log "github.com/iter8-tools/iter8/base/log"
)

// Experiment specification and result
type Experiment struct {
	// Tasks is the sequence of tasks that constitute this experiment
	Tasks  []TaskSpec        `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Result *ExperimentResult `json:"result,omitempty" yaml:"result,omitempty"`
}

// Task is an object that can be run
type Task interface {
	Run(exp *Experiment) error
	GetName() string
}

func GetIf(t Task) *string {
	var jsonBytes []byte
	var tm taskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to taskMeta
	_ = json.Unmarshal(jsonBytes, &tm)
	return tm.If
}

// ExperimentResult defines the current results from the experiment
type ExperimentResult struct {
	// StartTime is the time when the experiment run was started
	StartTime *time.Time `json:"startTime,omitempty" yaml:"startTime,omitempty"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// Failure is true if any of its tasks failed
	Failure bool `json:"failure" yaml:"failure"`

	// Insights produced in this experiment
	Insights *Insights `json:"insights,omitempty" yaml:"insights,omitempty"`
}

// Insights is a structure to contain experiment insights
type Insights struct {
	// NumAppVersions is the number of app versions detected by Iter8
	NumAppVersions *int `json:"numAppVersions,omitempty" yaml:"numVersions,omitempty"`

	// InsightInfo identifies the types of insights produced by this experiment
	InsightTypes []InsightType `json:"insightTypes,omitempty" yaml:"insightTypes,omitempty"`

	// MetricsInfo identifies the metrics involved in this experiment
	MetricsInfo map[string]MetricMeta `json:"metricsInfo,omitempty" yaml:"metricsInfo,omitempty"`

	// SLOStrs represents the SLOs involved in this experiment in string form
	SLOStrs []string `json:"SLOStrs,omitempty" yaml:"SLOStrs,omitempty"`

	// MetricValues:
	// the outer slice must be the same length as the number of app versions
	// the map key must match name of a metric in MetricsInfo
	// the inner slice contains the list of all observed metric values for given version and given metric; float value [i]["foo/bar"][k] is the [k]th observation for version [i] for the metric bar under backend foo.
	MetricValues []map[string][]float64 `json:"metricValues,omitempty" yaml:"metricValues,omitempty"`

	// SLOsSatisfied:
	// the outer slice must be of the same length as SLOStrs
	// the length of the inner slice must be the number of app versions
	// the boolean value at [i][j] indicate if SLO [i] is satisfied by version [j]
	SLOsSatisfied [][]bool `json:"SLOsSatisfied,omitempty" yaml:"SLOsSatisfied,omitempty"`

	// SLOsSatisfiedBy is the subset of versions that satisfy all SLOs
	// every integer in this slice must be in the range 0 to NumAppVersions - 1 (inclusive)
	SLOsSatisfiedBy []int `json:"SLOsSatisfiedBy,omitempty" yaml:"SLOsSatisfiedBy,omitempty"`
}

// InsightType identifies the type of insight
type InsightType string

const (
	// InsightTypeMetrics indicates metrics are observed during this experiment
	InsightTypeMetrics InsightType = "Metrics"

	// InsightTypeSLO indicatse SLOs are validated during this experiment
	InsightTypeSLO InsightType = "SLOs"
)

// MetricMeta describes a metric
type MetricMeta struct {
	Description string     `json:"description" yaml:"description"`
	Units       *string    `json:"units,omitempty" yaml:"units,omitempty"`
	Type        MetricType `json:"type" yaml:"type"`
}

// Criteria is list of criteria against which app versions are evaluated
type Criteria struct {
	// SLOs is a list of SLOs
	SLOs []SLO `json:"SLOs,omitempty" yaml:"SLOs,omitempty"`
}

// SLO is a service level objective
type SLO struct {
	// Metric is the fully qualified metric name (i.e., in the backendName/metricName format)
	Metric string `json:"metric" yaml:"metric"`

	// UpperLimit is the maximum acceptable value of the metric.
	UpperLimit *float64 `json:"upperLimit,omitempty" yaml:"upperLimit,omitempty"`

	// LowerLimit is the minimum acceptable value of the metric.
	LowerLimit *float64 `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty"`
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

// hasInsightType returns true if the experiment has a specific insight type set
func (in *Insights) hasInsightType(it InsightType) bool {
	if in != nil {
		if in.InsightTypes != nil {
			for _, v := range in.InsightTypes {
				if v == it {
					return true
				}
			}
		}
	}
	return false
}

// setInsightType adds a specific InsightType to the list of experiment insights types
func (in *Insights) setInsightType(it InsightType) error {
	if in.hasInsightType(it) {
		return nil
	}
	// LHS can be nil
	in.InsightTypes = append(in.InsightTypes, it)
	return nil
}

// setSLOStrs sets the SLOStrs field in insights
// if this function is called multiple times (example, due to looping), then
// it is intended to be called with the same argument each time
func (in *Insights) setSLOStrs(sloStrs []string) error {
	if in.SLOStrs != nil {
		if reflect.DeepEqual(in.SLOStrs, sloStrs) {
			return nil
		} else {
			log.Logger.WithStackTrace(fmt.Sprint("old: ", in.SLOStrs, "new: ", sloStrs)).Error("old and new value of sloStrs conflict")
			return errors.New("old and new value of sloStrs conflict")
		}
	}
	// LHS will be nil
	in.SLOStrs = sloStrs
	return nil
}

// initializeSLOsSatisfied initializes the SLOs satisfied field
func (e *Experiment) initializeSLOsSatisfied() error {
	if e.Result.Insights.SLOsSatisfied != nil {
		return nil // already initialized
	}
	// LHS will be nil
	e.Result.Insights.SLOsSatisfied = make([][]bool, len(e.Result.Insights.SLOStrs))
	for i := 0; i < len(e.Result.Insights.SLOStrs); i++ {
		e.Result.Insights.SLOsSatisfied[i] = make([]bool, *e.Result.Insights.NumAppVersions)
	}
	return nil
}

// initialize the number of app versions
func (in *Insights) initNumAppVersions(n int) error {
	if in.NumAppVersions != nil {
		if *in.NumAppVersions != n {
			errStr := fmt.Sprint("inconsistent number for app versions; old: ", *in.NumAppVersions, " new: ", n)
			log.Logger.Error(errStr)
			return errors.New(errStr)
		}
	}

	in.NumAppVersions = intPointer(n)
	return nil
}

// initialize metric values
func (in *Insights) initMetricValues(n int) error {
	if in.MetricValues != nil {
		if len(in.MetricValues) != n {
			errStr := fmt.Sprint("inconsistent number for app versions; in old metric values: ", len(in.MetricValues), " in new metric values: ", n)
			log.Logger.Error(errStr)
			return errors.New(errStr)
		} else {
			return nil
		}
	}

	in.MetricValues = make([]map[string][]float64, n)
	for i := 0; i < n; i++ {
		in.MetricValues[i] = make(map[string][]float64)
	}
	return nil
}

func (e *Experiment) InitResults() {
	e.Result = &ExperimentResult{
		StartTime:         nil,
		NumCompletedTasks: 0,
		Failure:           false,
		Insights: &Insights{
			NumAppVersions: nil,
			MetricsInfo:    map[string]MetricMeta{},
		},
	}
}

// SLOsBy returns true if version satisfies SLOs
func (exp *Experiment) SLOsBy(version int) bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Insights != nil {
				for _, v := range exp.Result.Insights.SLOsSatisfiedBy {
					if v == version {
						return true
					}
				}
			}
		}
	}
	return false
}

// SLOs returns true if all versions satisfy SLOs
func (exp *Experiment) SLOs() bool {
	if exp != nil {
		if exp.Result != nil {
			if exp.Result.Insights != nil {
				if exp.Result.Insights.NumAppVersions != nil {
					if exp.Result.Insights.SLOsSatisfiedBy != nil {
						if *exp.Result.Insights.NumAppVersions == len(exp.Result.Insights.SLOsSatisfiedBy) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
