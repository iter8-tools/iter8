package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeWrongTask(t *testing.T) {
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(CollectTaskName),
		},
		With: map[string]interface{}{
			"hello": "world",
		},
	}
	task, err := MakeAssess(ts)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestMakeAssess(t *testing.T) {
	// simple assess without any SLOs
	// should succeed
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(AssessTaskName),
		},
	}
	task, err := MakeAssess(ts)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	// incorrect with clause
	// should fail
	ts = &TaskSpec{
		taskMeta: taskMeta{Task: stringPointer(AssessTaskName)},
		With: map[string]interface{}{
			"SLOs": "hello world",
		},
	}
	task, err = MakeAssess(ts)
	assert.Error(t, err)
	assert.Nil(t, task)
}

// Test a runnable assert condition here
func TestRunAssess(t *testing.T) {
	// simple assess without any SLOs
	// should succeed
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(AssessTaskName),
		},
	}
	task, _ := MakeAssess(ts)
	exp := &Experiment{
		Tasks: []TaskSpec{},
	}
	exp.InitResults()
	task.Run(exp)

	// assess with an SLO
	// should succeed
	ts = &TaskSpec{
		taskMeta: taskMeta{Task: stringPointer(AssessTaskName)},
		With: map[string]interface{}{
			"SLOs": []SLO{{
				Metric:     "m",
				UpperLimit: float64Pointer(20.0),
			}},
		},
	}
	task, _ = MakeAssess(ts)
	exp = &Experiment{
		Tasks:  []TaskSpec{},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.InitInsights(1, []InsightType{InsightTypeMetrics})
	err := task.Run(exp)
	assert.NoError(t, err)

	// assess with an experiment where num versions is 1
	exp.Result.Insights.NumVersions = 1
	err = task.Run(exp)
	assert.NoError(t, err)

}
