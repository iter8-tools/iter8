package base

import (
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRunCollectHTTP(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	// Exact URL match
	httpmock.RegisterResponder("POST", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	// valid collect HTTP task... should succeed
	endpoints := map[string]collectHTTPInputsHelper{}
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			collectHTTPInputsHelper{
				Duration:    StringPointer("1s"),
				PayloadFile: StringPointer(CompletePath("../", "testdata/payload/ukpolice.json")),
				Headers:     map[string]string{},
				URL:         "https://something.com",
			},
			endpoints,
		},
	}

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err := ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)
}
