package base

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRunCollect(t *testing.T) {
	// valid collect task... should succeed
	ct := &collectTask{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
		With: collectInputs{
			Duration: StringPointer("1s"),
			VersionInfo: []*version{{
				Headers: map[string]string{},
				URL:     "https://something.com",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err := ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
}
