package base

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeWrongRunTask(t *testing.T) {
	// no run ...
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(AssessTaskName),
		},
		With: map[string]interface{}{
			"hello": "world",
		},
	}
	task, err := MakeRun(ts)
	assert.Error(t, err)
	assert.Nil(t, task)

	// both run and task
	ts = &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(AssessTaskName),
			Run:  stringPointer("echo hello"),
		},
		With: map[string]interface{}{
			"hello": "world",
		},
	}
	task, err = MakeRun(ts)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestMakeRun(t *testing.T) {
	// valid run task... should succeed
	rt := &runTask{
		taskMeta: taskMeta{
			Run: stringPointer("echo hello"),
		},
	}

	ts := &TaskSpec{}
	b, _ := json.Marshal(rt)
	json.Unmarshal(b, ts)
	task, err := MakeRun(ts)
	assert.NoError(t, err)
	assert.NotNil(t, task)
}

func TestRunRun(t *testing.T) {
	// valid collect task... should succeed
	rt := &runTask{
		taskMeta: taskMeta{
			Run: stringPointer("echo hello"),
		},
	}

	ts := &TaskSpec{}
	b, _ := json.Marshal(rt)
	json.Unmarshal(b, ts)
	task, err := MakeRun(ts)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	exp := &Experiment{
		Tasks:  []TaskSpec{*ts},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err = rt.Run(exp)
	assert.NoError(t, err)
}
