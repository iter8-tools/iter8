package base

import (
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestMakeWrongCollectTask(t *testing.T) {
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: map[string]interface{}{
			"hello": "world",
		},
	}
	task, err := MakeCollect(ts)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestMakeCollect(t *testing.T) {
	// collect without version info ... should fail
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
	}
	task, err := MakeCollect(ts)
	assert.Error(t, err)
	assert.Nil(t, task)

	// collect task with only nil versions... should fail
	ct := &collectTask{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
		With: collectInputs{
			VersionInfo: []*version{nil, nil},
		},
	}

	b, _ := json.Marshal(ct)
	json.Unmarshal(b, ts)
	task, err = MakeCollect(ts)
	assert.Error(t, err)
	assert.Nil(t, task)

	// valid collect task... should succeed
	ct = &collectTask{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
		With: collectInputs{
			VersionInfo: []*version{nil, {
				Headers: map[string]string{},
				URL:     "https://something.com",
			}},
		},
	}

	b, _ = json.Marshal(ct)
	json.Unmarshal(b, ts)
	task, err = MakeCollect(ts)
	assert.NoError(t, err)
	assert.NotNil(t, task)
}

func TestRunCollect(t *testing.T) {
	// valid collect task... should succeed
	ct := &collectTask{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
		With: collectInputs{
			VersionInfo: []*version{{
				Headers: map[string]string{},
				URL:     "https://something.com",
			}},
		},
	}

	ts := &TaskSpec{}
	b, err := json.Marshal(ct)
	assert.NoError(t, err)
	err = json.Unmarshal(b, ts)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	exp := &Experiment{
		Tasks:  []TaskSpec{*ts},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err = ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
}
