package base

import (
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRunExperiment(t *testing.T) {
	// valid collect task... should succeed
	ct := &collectTask{
		taskMeta: taskMeta{
			Task: StringPointer(CollectTaskName),
		},
		With: collectInputs{
			Duration:    StringPointer("1s"),
			VersionInfo: []*version{{Headers: map[string]string{}, URL: "https://something.com"}},
		},
	}

	tsc := &TaskSpec{}
	b, err := json.Marshal(ct)
	assert.NoError(t, err)
	err = json.Unmarshal(b, tsc)
	assert.NoError(t, err)

	// valid assess task... should succeed
	at := &assessTask{
		taskMeta: taskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     iter8BuiltInPrefix + "/" + errorCountMetricName,
				UpperLimit: float64Pointer(0),
			}},
		},
	}

	tsa := &TaskSpec{}
	b, err = json.Marshal(at)
	assert.NoError(t, err)
	err = json.Unmarshal(b, tsa)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	exp := &Experiment{
		Tasks:  []TaskSpec{*tsc, *tsa},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err = ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	// experiment should contain histogram metrics
	assert.True(t, exp.ContainsInsight(InsightTypeHistMetrics))

	// SLOs should be satisfied by app
	assert.True(t, exp.SLOs())
}
