package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test a runnable assert condition here
func TestRunAssess(t *testing.T) {
	// simple assess without any SLOs
	// should succeed
	task := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{},
	}
	exp := &Experiment{
		Tasks: []Task{task},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)
	err := task.run(exp)
	assert.NoError(t, err)

	// assess with an SLO
	// should succeed
	task.With = assessInputs{
		SLOs: []SLO{{
			Metric:     "a/b",
			UpperLimit: float64Pointer(20.0),
		}},
	}
	err = task.run(exp)
	assert.NoError(t, err)
}
