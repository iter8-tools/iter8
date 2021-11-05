package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
		return nil, errors.New("task need to be " + AssessTaskName)
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

// Run executes the assess-versions task
func (t *assessTask) Run(exp *Experiment) error {
	err := exp.setTestingPattern(t.With.Criteria)
	if err != nil {
		return err
	}

	err = exp.setObjectives(evaluateObjectives(exp, t.With.Criteria.Objectives))
	if err != nil {
		return err
	}

	err = exp.setSatisfying(computeSatisfying(exp))
	if err != nil {
		return err
	}

	err = exp.setWinner(findWinner(exp))
	if err != nil {
		return err
	}

	return err
}

// compute set of versions satisfying objectives
func computeSatisfying(exp *Experiment) []string {
	satisfying := []string{}
	if exp.Result == nil || exp.Result.NumAppVersions == nil {
		log.Logger.Warn("unknown number of app versions")
		return satisfying
	}
	for i := 0; i < *exp.Result.NumAppVersions; i++ {
		satisfied := true
		for j := range exp.Result.Analysis.Objectives {
			satisfied = satisfied && exp.Result.Analysis.Objectives[i][j]
		}
		if satisfied {
			satisfying = append(satisfying, "v"+fmt.Sprint(i))
		}
	}
	return satisfying
}

// evaluate objectives
func evaluateObjectives(exp *Experiment, objs []Objective) [][]bool {
	if exp.Result == nil || exp.Result.NumAppVersions == nil {
		log.Logger.Warn("unknown number of app versions")
		return nil
	}
	if *exp.Result.Analysis.TestingPattern == TestingPatternSLOValidation {
		// Objectives
		// if not empty, the length of the outer slice must match the length of Spec.Versions
		// if not empty, the length of an inner slice must match the number of objectives in the assess-versions task
		objAssessment := make([][]bool, *exp.Result.NumAppVersions)
		for i := 0; i < *exp.Result.NumAppVersions; i++ {
			objAssessment[i] = make([]bool, len(objs))
			for j := range objs {
				objAssessment[i][j] = objectiveSatisfied(exp, i, objs[j])
			}
		}
		return objAssessment
	} else {
		return nil
	}
}

// return true if version i satisfies objective j
func objectiveSatisfied(e *Experiment, i int, o Objective) bool {
	// get metric value
	val := getMetricValue(e, i, o.Metric)
	// check if metric is available
	if val == nil {
		log.Logger.Warn(fmt.Sprintf("unable to find value for version %v and metric %s", i, o.Metric))
		return false
	}
	// check lower and upper limits
	if o.LowerLimit != nil {
		if *val < *o.LowerLimit {
			return false
		}
	}
	if o.UpperLimit != nil {
		if *val > *o.UpperLimit {
			return false
		}
	}
	return true
}

// get the value of the given metric for the given version
func getMetricValue(e *Experiment, i int, m string) *float64 {
	if !strings.HasPrefix(m, iter8FortioPrefix) {
		log.Logger.Warn("unknown backend detected in metric " + m)
		return nil
	}

	if e == nil || e.Result == nil || e.Result.Analysis == nil || e.Result.Analysis.Metrics == nil {
		log.Logger.Warn("metrics unavailable in experiment")
		return nil
	}

	if e.Result.NumAppVersions == nil {
		log.Logger.Warn("unknown number of app versions")
		return nil
	}

	if len(e.Result.Analysis.Metrics) != *e.Result.NumAppVersions {
		log.Logger.Warn("metrics slice must be of the same length as versions slice")
		return nil
	}

	if e.Result.Analysis.Metrics[i] == nil {
		log.Logger.Warn("no metrics available for version ", i)
		return nil
	}

	if vals, ok := e.Result.Analysis.Metrics[i][m]; !ok || len(vals) == 0 {
		log.Logger.Warn("metric ", m, " unavailable for version ", i)
		return nil
	} else {
		return float64Pointer(vals[len(vals)-1])
	}
}

// find winning version
func findWinner(exp *Experiment) *string {
	if exp.Result.NumAppVersions == nil {
		log.Logger.Warn("unknown number of app versions")
		return nil
	}

	if *exp.Result.Analysis.TestingPattern == TestingPatternSLOValidation {
		if *exp.Result.NumAppVersions == 1 {
			// check if all objectives are satisfied
			for i, sat := range exp.Result.Analysis.Objectives[0] {
				if !sat {
					log.Logger.Info("version 0 failed to satisfy objective " + fmt.Sprintf("%v", i))
					return nil
				}
			}
			log.Logger.Info("all objectives satisfied by winning version 0")
			return stringPointer("v0")
		} else {
			log.Logger.Warn("winner with multiple versions undefined for testing pattern " + TestingPatternSLOValidation)
		}
	} else {
		log.Logger.Warn("winner undefined for testing pattern " + string(*exp.Result.Analysis.TestingPattern))
	}
	return nil
}
