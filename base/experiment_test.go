package base

import (
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/iter8/base/log"
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

	b, err = ioutil.ReadFile(CompletePath("../testdata", "experiment_db.yaml"))
	assert.NoError(t, err)
	es = &ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(*es))
}
func TestRunTask(t *testing.T) {
	SetupWithMock(t)

	// valid collect task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration: StringPointer("1s"),
			Headers:  map[string]string{},
			URL:      "https://httpbin.org/get",
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
		Spec:   []Task{ct, at},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err := ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	// SLOs should be satisfied by app
	for i := 0; i < len(exp.Result.Insights.SLOs); i++ { // i^th SLO
		assert.True(t, exp.Result.Insights.SLOsSatisfied[i][0]) // satisfied by only version
	}
}

func TestRunExperiment(t *testing.T) {
	SetupWithMock(t)
	b, err := ioutil.ReadFile(CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	es := &ExperimentSpec{}
	err = yaml.Unmarshal(b, es)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(*es))

	exp := Experiment{
		Spec: *es,
	}

	err = RunExperiment(false, &mockDriver{&exp})
	assert.NoError(t, err)

	yamlBytes, _ := yaml.Marshal(exp.Result)
	log.Logger.WithStackTrace(string(yamlBytes)).Debug("results")
	assert.True(t, exp.Completed())
	assert.True(t, exp.NoFailure())
	expRes, _ := yaml.Marshal(exp.Result)
	log.Logger.Debug(string(expRes))
	assert.True(t, exp.SLOs())

}

func TestFailExperiment(t *testing.T) {
	exp := Experiment{
		Spec: ExperimentSpec{},
	}
	exp.initResults(1)

	exp.failExperiment()
	assert.False(t, exp.NoFailure())
}
