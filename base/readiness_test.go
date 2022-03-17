package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidName(t *testing.T) {
	// create task
	rTask := &ReadinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: ReadinessInputs{
			Objects: []ObjRef{{
				Name: "valid-name",
			}},
		},
	}
	rTask.initializeDefaults()
	err := rTask.validateInputs()
	assert.NoError(t, err)
}
func TestInvalidName(t *testing.T) {
	// create task
	rTask := &ReadinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: ReadinessInputs{
			Objects: []ObjRef{{
				Name: "invalid#name--has#hash",
			}},
		},
	}
	rTask.initializeDefaults()
	err := rTask.validateInputs()
	assert.Error(t, err)
}

func TestLongName(t *testing.T) {
	// create task
	rTask := &ReadinessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(ReadinessTaskName),
		},
		With: ReadinessInputs{
			Objects: []ObjRef{{
				Name: "name-that-is-too-long-because-it-is-more-than-sixty-three-characters",
			}},
		},
	}
	rTask.initializeDefaults()
	err := rTask.validateInputs()
	assert.Error(t, err)

	// // create experiment
	// exp := &Experiment{
	// 	Tasks:  []Task{rTask},
	// 	Result: &ExperimentResult{},
	// }
	// exp.InitResults()

	// // run task
	// err := rTask.Run(exp)
	// assert.Error(t, err)
}
