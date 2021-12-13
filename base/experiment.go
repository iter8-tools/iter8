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
	Tasks  []TaskSpec        `json:"tasks" yaml:"tasks"`
	Result *ExperimentResult `json:"result" yaml:"result"`
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
	// StartTime is the time when the experiment run started
	StartTime time.Time `json:"startTime" yaml:"startTime"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// Failure is true if any of its tasks failed
	Failure bool `json:"failure" yaml:"failure"`

	// Insights produced in this experiment
	Insights *Insights `json:"insights,omitempty" yaml:"insights,omitempty"`
}

// Insights is a structure to contain experiment insights
type Insights struct {
	// NumVersions is the number of app versions detected by Iter8
	NumVersions int `json:"numVersions" yaml:"numVersions"`

	// InsightInfo identifies the types of insights produced by this experiment
	InsightTypes []InsightType `json:"insightTypes,omitempty" yaml:"insightTypes,omitempty"`

	// MetricsInfo identifies the metrics involved in this experiment
	MetricsInfo map[string]MetricMeta `json:"metricsInfo,omitempty" yaml:"metricsInfo,omitempty"`

	// SLOs involved in this experiment
	SLOs []SLO `json:"SLOs,omitempty" yaml:"SLOs,omitempty"`

	// MetricValues:
	// the outer slice must be the same length as the number of app versions
	// the map key must match name of a metric in MetricsInfo
	// the inner slice contains the list of all observed metric values for given version and given metric; float value [i]["foo/bar"][k] is the [k]th observation for version [i] for the metric bar under backend foo.
	MetricValues []map[string][]float64 `json:"metricValues,omitempty" yaml:"metricValues,omitempty"`

	// SLOsSatisfied:
	// the outer slice must be of the same length as SLOs
	// the length of the inner slice must be the number of app versions
	// the boolean value at [i][j] indicate if SLO [i] is satisfied by version [j]
	SLOsSatisfied [][]bool `json:"SLOsSatisfied,omitempty" yaml:"SLOsSatisfied,omitempty"`

	// SLOsSatisfiedBy is the subset of versions that satisfy all SLOs
	// every integer in this slice must be in the range 0 to NumVersions - 1 (inclusive)
	SLOsSatisfiedBy []int `json:"SLOsSatisfiedBy,omitempty" yaml:"SLOsSatisfiedBy,omitempty"`
}

// InsightType identifies the type of insight
type InsightType string

const (
	// InsightTypeHistMetrics indicates histogram metrics are collected during this experiment
	InsightTypeHistMetrics InsightType = "HistMetrics"

	// InsightTypeMetrics indicates metrics are collected during this experiment
	InsightTypeMetrics InsightType = "Metrics"

	// InsightTypeSLO indicatse SLOs are validated during this experiment
	InsightTypeSLO InsightType = "SLOs"
)

// MetricMeta describes a metric
type MetricMeta struct {
	Description string     `json:"description" yaml:"description"`
	Units       *string    `json:"units,omitempty" yaml:"units,omitempty"`
	Type        MetricType `json:"type" yaml:"type"`
	XMin        *float64   `json:"xmin" yaml:"xmin"`
	XMax        *float64   `json:"xmax" yaml:"xmax"`
	NumBuckets  *int       `json:"numBuckets" yaml:"numBuckets"`
}

// SLO is a service level objective
type SLO struct {
	// Metric is the fully qualified metric name (i.e., in the backendName/metricName format)
	Metric string `json:"metric" yaml:"metric" validate:"gt=0,required"`

	// UpperLimit is the maximum acceptable value of the metric.
	UpperLimit *float64 `json:"upperLimit,omitempty" yaml:"upperLimit,omitempty" validate:"required_without=LowerLimit"`

	// LowerLimit is the minimum acceptable value of the metric.
	LowerLimit *float64 `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty" validate:"required_without=UpperLimit"`
}

type taskMeta struct {
	Task *string `json:"task,omitempty" yaml:"task,omitempty" validate:"required_without=Run"`
	Run  *string `json:"run,omitempty" yaml:"run,omitempty" validate:"required_without=Task"`
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
func (in *Insights) setInsightType(it InsightType) {
	if !in.hasInsightType(it) {
		in.InsightTypes = append(in.InsightTypes, it)
	}
}

// setSLOs sets the SLOs field in insights
// if this function is called multiple times (example, due to looping), then
// it is intended to be called with the same argument each time
func (in *Insights) setSLOs(slos []SLO) error {
	if in.SLOs != nil {
		if reflect.DeepEqual(in.SLOs, slos) {
			return nil
		} else {
			e := fmt.Errorf("old and new value of slos conflict")
			log.Logger.WithStackTrace(fmt.Sprint("old: ", in.SLOs, "new: ", slos)).Error(e)
			return e
		}
	}
	// LHS will be nil
	in.SLOs = slos
	return nil
}

// initializeSLOsSatisfied initializes the SLOs satisfied field
func (e *Experiment) initializeSLOsSatisfied() error {
	if e.Result.Insights.SLOsSatisfied != nil {
		return nil // already initialized
	}
	// LHS will be nil
	e.Result.Insights.SLOsSatisfied = make([][]bool, len(e.Result.Insights.SLOs))
	for i := 0; i < len(e.Result.Insights.SLOs); i++ {
		e.Result.Insights.SLOsSatisfied[i] = make([]bool, e.Result.Insights.NumVersions)
	}
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
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
	}
}

func (r *ExperimentResult) InitInsights(n int, it []InsightType) {
	r.Insights = &Insights{
		NumVersions:  n,
		InsightTypes: it,
		MetricsInfo:  make(map[string]MetricMeta),
		MetricValues: make([]map[string][]float64, n),
	}
	for i := 0; i < n; i++ {
		r.Insights.MetricValues[i] = make(map[string][]float64)
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
				if exp.Result.Insights.SLOsSatisfiedBy != nil {
					if exp.Result.Insights.NumVersions == len(exp.Result.Insights.SLOsSatisfiedBy) {
						return true
					}
				}
			}
		}
	}
	return false
}
