package base

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRunCollectHTTP(t *testing.T) {
	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration:   StringPointer("1s"),
			PayloadURL: StringPointer("https://data.police.uk/api/crimes-street-dates"),
			VersionInfo: []*versionHTTP{{
				Headers: map[string]string{},
				URL:     "https://something.com",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("POST", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	httpmock.RegisterResponder("GET", "https://data.police.uk/api/crimes-street-dates",
		httpmock.NewStringResponder(200, `[{"my": 1, "great": "payload"}]`))

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err := ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
}
