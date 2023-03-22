package base

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test a runnable assert condition here
func TestRunAssess(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	// simple assess without any SLOs
	// should succeed
	task := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{},
	}
	exp := &Experiment{
		Spec: []Task{task},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err := task.run(exp)
	assert.NoError(t, err)

	// assess with an SLO
	// should succeed
	task.With = assessInputs{
		SLOs: &SLOLimits{
			Upper: []SLO{{
				Metric: "a/b",
				Limit:  20.0,
			}},
		},
		Rewards: &Rewards{
			Max: []string{"a/b"},
		},
	}
	err = task.run(exp)
	assert.NoError(t, err)
}
