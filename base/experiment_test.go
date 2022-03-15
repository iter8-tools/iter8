package base

import (
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base/log"
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
func TestRunTask(t *testing.T) {
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
				Metric:     httpMetricPrefix + "/" + builtInHTTPErrorCountId,
				UpperLimit: float64Pointer(0),
			}},
		},
	}

	exp := &Experiment{
		Tasks:  []Task{ct, at},
		Result: &ExperimentResult{},
	}
	exp.initResults()
	err := ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	// SLOs should be satisfied by app
	for i := 0; i < len(exp.Result.Insights.SLOs); i++ { // i^th SLO
		assert.True(t, exp.Result.Insights.SLOsSatisfied[i][0]) // satisfied by only version
	}

	httpmock.DeactivateAndReset()

}

type mockDriver struct {
	*Experiment
}

func (m *mockDriver) ReadResult() (*ExperimentResult, error) {
	return m.Experiment.Result, nil
}

func (m *mockDriver) WriteResult(r *ExperimentResult) error {
	m.Experiment.Result = r
	return nil
}

func (m *mockDriver) ReadSpec() (ExperimentSpec, error) {
	return m.Experiment.Tasks, nil
}

func TestRunExperiment(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	b, err := ioutil.ReadFile(CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	es := &ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(*es))

	exp := Experiment{
		Tasks: *es,
	}

	err = RunExperiment(&mockDriver{&exp})
	assert.NoError(t, err)

	yamlBytes, _ := yaml.Marshal(exp.Result)
	log.Logger.WithStackTrace(string(yamlBytes)).Debug("results")
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	expRes, _ := yaml.Marshal(exp.Result)
	log.Logger.Info(string(expRes))
	assert.True(t, exp.SLOs())

	httpmock.DeactivateAndReset()
}

func TestFailExperiment(t *testing.T) {
	exp := Experiment{
		Tasks: ExperimentSpec{},
	}
	exp.initResults()

	exp.failExperiment()
	assert.False(t, exp.NoFailure())
}
