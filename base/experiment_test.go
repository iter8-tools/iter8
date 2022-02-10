package base

import (
	"io/ioutil"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestReadExperiment(t *testing.T) {
	b, err := ioutil.ReadFile(CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	es := &ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(*es))

	b, err = ioutil.ReadFile(CompletePath("../testdata", "experiment_grpc.yaml"))
	assert.NoError(t, err)
	es = &ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(*es))
}
func TestRunExperiment(t *testing.T) {
	httpmock.Activate()

	// Exact URL match
	httpmock.RegisterResponder("GET", "https://something.com",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	// valid collect task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration:    StringPointer("1s"),
			VersionInfo: []*versionHTTP{{Headers: map[string]string{}, URL: "https://something.com"}},
		},
	}

	// valid assess task... should succeed
	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     iter8BuiltInPrefix + "/" + builtInHTTPErrorCountId,
				UpperLimit: float64Pointer(0),
			}},
		},
	}

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

	httpmock.DeactivateAndReset()
}
