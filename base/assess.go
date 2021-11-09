package base

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iter8-tools/iter8/base/log"
)

// assessInputs contain the inputs to the assess-versions task to be executed.
type assessInputs struct {
	// Criteria is the assessment criteria
	Criteria *Criteria `json:"criteria" yaml:"criteria"`
}

// assessTask enables assessment of versions
type assessTask struct {
	taskMeta
	With assessInputs `json:"with" yaml:"with"`
}

const (
	// AssessTaskName is the name of the task this file implements
	AssessTaskName = "assess-versions"
)

// MakeAssess constructs an asessTask out of a task spec
func MakeAssess(t *TaskSpec) (Task, error) {
	if t == nil || t.Task == nil || *t.Task != AssessTaskName {
		return nil, errors.New("task needs to be " + AssessTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt Task
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonString to CollectTask
	ct := &assessTask{}
	err = json.Unmarshal(jsonBytes, &ct)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to unmarshal assess task")
		return nil, err
	}
	bt = ct
	return bt, nil
}

// get string representation of SLOs
func getSLOStrs(slos []SLO) []string {
	sloStrs := []string{}
	for _, v := range slos {
		str := ""
		if v.LowerLimit != nil {
			str += fmt.Sprint(*v.LowerLimit, " <= ")
		}
		str += v.Metric
		if v.UpperLimit != nil {
			str += fmt.Sprint(" <= ", *v.UpperLimit)
		}
		sloStrs = append(sloStrs, str)
	}
	return sloStrs
}

// Run executes the assess-versions task
func (t *assessTask) Run(exp *Experiment) error {
	if t.With.Criteria == nil ||
		t.With.Criteria.SLOs == nil ||
		len(t.With.Criteria.SLOs) == 0 ||
		exp.Result.Insights.NumAppVersions == nil ||
		*exp.Result.Insights.NumAppVersions == 0 {
		// do nothing for now
		// todo: fix when rewards are introduced

		log.Logger.Warn("nothing to do; returning")
		return nil
	}

	// set insight type (if needed)
	err := exp.Result.Insights.setInsightType(InsightTypeSLO)
	if err != nil {
		return err
	}

	// set SLOStrs (if needed)
	err = exp.Result.Insights.setSLOStrs(getSLOStrs(t.With.Criteria.SLOs))
	if err != nil {
		return err
	}

	// set initialize SLOsSatisfied (if needed)
	err = exp.initializeSLOsSatisfied()
	if err != nil {
		return err
	}

	// set SLOsSatisfied
	exp.Result.Insights.SLOsSatisfied = evaluateSLOs(exp, t.With.Criteria.SLOs)

	// set SLOsSatisfiedBy
	exp.Result.Insights.SLOsSatisfiedBy = computeSLOsSatisfiedBy(exp)

	return err
}

// evaluate SLOs
func evaluateSLOs(exp *Experiment, slos []SLO) [][]bool {
	slosSatisfied := make([][]bool, len(slos))
	for i := 0; i < len(slos); i++ {
		slosSatisfied[i] = make([]bool, *exp.Result.Insights.NumAppVersions)
		for j := 0; j < *exp.Result.Insights.NumAppVersions; j++ {
			slosSatisfied[i][j] = sloSatisfied(exp, slos, i, j)
		}
	}
	return slosSatisfied
}

// return true if SLO i satisfied by version j
func sloSatisfied(e *Experiment, slos []SLO, i int, j int) bool {
	val := getMetricValue(e, j, slos[i].Metric)
	// check if metric is available
	if val == nil {
		log.Logger.Warnf("unable to find value for version %v and metric %s", j, slos[i].Metric)
		return false
	}
	// check lower limit
	if slos[i].LowerLimit != nil {
		if *val < *slos[i].LowerLimit {
			return false
		}
	}
	// check upper limit
	if slos[i].UpperLimit != nil {
		if *val > *slos[i].UpperLimit {
			return false
		}
	}
	return true
}

// computeSLOsSatisfiedBy computes the subset of versions that satisfy SLOs
func computeSLOsSatisfiedBy(exp *Experiment) []int {
	sats := []int{}
	for j := 0; j < *exp.Result.Insights.NumAppVersions; j++ {
		sat := true
		for i := range exp.Result.Insights.SLOStrs {
			sat = sat && exp.Result.Insights.SLOsSatisfied[i][j]
		}
		if sat {
			sats = append(sats, j)
		}
	}
	return sats
}

// get the value of the given metric for the given version
func getMetricValue(e *Experiment, i int, m string) *float64 {
	vals := e.Result.Insights.MetricValues[i][m]
	if len(vals) == 0 {
		return nil
	}
	return float64Pointer(vals[len(vals)-1])
}
