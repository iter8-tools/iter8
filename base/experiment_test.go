package base

import (
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func TestReadExperiment(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	b, err := os.ReadFile(CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	e := &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(e.Spec))

	b, err = os.ReadFile(CompletePath("../testdata", "experiment_grpc.yaml"))
	assert.NoError(t, err)
	e = &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(e.Spec))

	b, err = os.ReadFile(CompletePath("../testdata", "experiment_db.yaml"))
	assert.NoError(t, err)
	e = &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(e.Spec))
}

func TestRunningTasks(t *testing.T) {
	_ = os.Chdir(t.TempDir())
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
			SLOs: &SLOLimits{
				Upper: []SLO{{
					Metric: httpMetricPrefix + "/" + builtInHTTPErrorCountID,
					Limit:  0,
				}},
			},
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

	err = at.run(exp)
	assert.NoError(t, err)

	// SLOs should be satisfied by app
	for i := 0; i < len(exp.Result.Insights.SLOs.Upper); i++ { // i^th SLO
		assert.True(t, exp.Result.Insights.SLOsSatisfied.Upper[i][0]) // satisfied by only version
	}
}

func TestRunExperiment(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	SetupWithMock(t)

	b, err := os.ReadFile(CompletePath("../testdata", "experiment.yaml"))
	assert.NoError(t, err)
	e := &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(e.Spec))

	err = RunExperiment(false, &mockDriver{e})
	assert.NoError(t, err)

	assert.True(t, e.Completed())
	assert.True(t, e.NoFailure())
	expBytes, _ := yaml.Marshal(e)
	log.Logger.Debug("\n" + string(expBytes))
	assert.True(t, e.SLOs())
}

func TestFailExperiment(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	exp := Experiment{
		Spec: ExperimentSpec{},
	}
	exp.initResults(1)

	exp.failExperiment()
	assert.False(t, exp.NoFailure())
}
