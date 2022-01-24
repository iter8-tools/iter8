package base

import (
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

	// valid assess task... should succeed
	at := &assessTask{
		taskMeta: taskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     iter8BuiltInPrefix + "/" + builtInHTTPErrorCountId,
				UpperLimit: float64Pointer(0),
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Exact URL match
	httpmock.RegisterResponder("GET", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	exp := &Experiment{
		Tasks:  []Task{ct, at},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err := ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	// SLOs should be satisfied by app
	for i := 0; i < len(exp.Result.Insights.SLOs); i++ { // i^th SLO
		assert.True(t, exp.Result.Insights.SLOsSatisfied[i][0]) // satisfied by only version
	}
}