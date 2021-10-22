package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/core"
)

const (
	// AssessTaskName is the name of the task this file implements
	AssessTaskName string = "assess-versions"
)

// AssessInputs contain the inputs to the assess-versions task to be executed.
type AssessInputs struct {
	// Criteria is the assessment criteria
	Criteria *core.Criteria `json:"criteria,omitempty" yaml:"criteria,omitempty"`
}

// AssessTask enables assessment of versions
type AssessTask struct {
	core.TaskMeta
	With AssessInputs `json:"with" yaml:"with"`
}

// MakeAssess constructs a AssessTask out of a assess versions task spec
func MakeAssess(t *core.TaskSpec) (core.Task, error) {
	if *t.Task != AssessTaskName {
		return nil, errors.New("task need to be " + AssessTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt core.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to CollectTask
	if err == nil {
		ct := &AssessTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		bt = ct
	}
	return bt, err
}

// Run executes the assess-versions task
func (t *AssessTask) Run(exp *core.Experiment) error {
	err := exp.SetTestingPattern(t.With.Criteria)
	if err != nil {
		return err
	}

	err = exp.SetObjectives(evaluateObjectives(exp, t.With.Criteria.Objectives))
	if err != nil {
		return err
	}

	err = exp.SetWinner(findWinner(exp))
	if err != nil {
		return err
	}

	return err
}

// evaluate objectives
func evaluateObjectives(exp *core.Experiment, objs []core.Objective) [][]bool {
	if *exp.Result.Analysis.TestingPattern == core.TestingPatternSLOValidation {
		// Objectives
		// if not empty, the length of the outer slice must match the length of Spec.Versions
		// if not empty, the length of an inner slice must match the number of objectives in the assess-versions task
		objAssessment := make([][]bool, len(exp.Spec.Versions))
		for i := range exp.Spec.Versions {
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
func objectiveSatisfied(e *core.Experiment, i int, o core.Objective) bool {
	// get metric value
	val := getMetricValue(e, i, o.Metric)
	// what kind of objective is this
	if val == nil {
		core.Logger.Warn(fmt.Sprintf("unable to find value for version %s and metric %s", e.Spec.Versions[i], o.Metric))
		return false
	}
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
func getMetricValue(e *core.Experiment, i int, m string) *float64 {
	if !strings.HasPrefix(m, core.IFBackend.Name) {
		core.Logger.Warn("unknown backend detected in metric " + m)
		return nil
	}

	if !core.IFBackend.HasMetric(m) {
		core.Logger.Warn("unknown metric " + m + " detected in backend " + core.IFBackend.Name)
		return nil
	}

	if e == nil || e.Result == nil || e.Result.Analysis == nil || e.Result.Analysis.Metrics == nil {
		core.Logger.Warn("metrics unavailable in experiment")
		return nil
	}

	if len(e.Result.Analysis.Metrics) != len(e.Spec.Versions) {
		core.Logger.Warn("metrics slice must be of the same length as versions slice")
		return nil
	}

	if e.Result.Analysis.Metrics[i] == nil {
		core.Logger.Warn("no metrics available for version " + e.Spec.Versions[i])
		return nil
	}

	if vals, ok := e.Result.Analysis.Metrics[i][m]; !ok {
		core.Logger.Warn("metrics unavailable for version " + e.Spec.Versions[i])
		return nil
	} else if len(vals) == 0 {
		core.Logger.Warn("metric " + m + "unavailable for version " + e.Spec.Versions[i])
		return nil
	} else {
		return core.Float64Pointer(vals[len(vals)-1])
	}

}

// find winning version
func findWinner(exp *core.Experiment) *string {
	if *exp.Result.Analysis.TestingPattern == core.TestingPatternSLOValidation {
		if len(exp.Spec.Versions) == 1 {
			// check if all objectives are satisfied
			for i, sat := range exp.Result.Analysis.Objectives[0] {
				if !sat {
					core.Logger.Info("version " + exp.Spec.Versions[0] + " failed to satisfy objective " + fmt.Sprintf("%v", i))
					return nil
				}
			}
			core.Logger.Info("all objectives satisfied by winner " + exp.Spec.Versions[0])
			return &exp.Spec.Versions[0]
		} else {
			core.Logger.Warn("winner with multiple versions undefined for testing pattern " + core.TestingPatternSLOValidation)
		}
	} else {
		core.Logger.Warn("winner undefined for testing pattern " + string(*exp.Result.Analysis.TestingPattern))
	}
	return nil
}
