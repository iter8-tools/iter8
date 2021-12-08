package base

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeWrongCollectTask(t *testing.T) {
	ts := &TaskSpec{
		taskMeta: taskMeta{
			Task: stringPointer(AssessTaskName),
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
			Task: stringPointer(CollectTaskName),
		},
	}
	task, err := MakeCollect(ts)
	assert.Error(t, err)
	assert.Nil(t, task)

	// collect task with only nil versions... should fail
	ct := &collectTask{
		taskMeta: taskMeta{
			Task: stringPointer(CollectTaskName),
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
			Task: stringPointer(CollectTaskName),
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
