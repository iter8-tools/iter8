package base

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

type readScenario struct {
	good  string
	tasks int
	bad   string
}

func TestReadExp(t *testing.T) {
	testcases := map[string]readScenario{
		"http":       {good: "experiment.yaml", tasks: 4, bad: ""},
		"grpc":       {good: "experiment_grpc.yaml", tasks: 3, bad: ""},
		"custom db ": {good: "experiment_db.yaml", tasks: 4, bad: ""},
		"abn task":   {good: "experiment_abn.yaml", tasks: 1, bad: "experiment_abn_bad.yaml"},
	}
	for label, s := range testcases {
		os.Chdir(t.TempDir())
		t.Run(label, func(t *testing.T) {
			testReadExp(t, s)
		})
	}
}
func testReadExp(t *testing.T, s readScenario) {
	// valid experiment yaml
	b, err := ioutil.ReadFile(CompletePath("../testdata", s.good))
	assert.NoError(t, err)
	e := &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, s.tasks, len(e.Spec))
	if s.bad != "" {
		// invalid experiment yaml
		b, err = ioutil.ReadFile(CompletePath("../testdata", s.bad))
		assert.NoError(t, err)
		e = &Experiment{}
		err = yaml.Unmarshal(b, e)
		assert.Error(t, err)
	}
}

func TestRunningTasks(t *testing.T) {
	os.Chdir(t.TempDir())
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
					Metric: httpMetricPrefix + "/" + builtInHTTPErrorCountId,
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
	os.Chdir(t.TempDir())
	SetupWithMock(t)

	b, err := ioutil.ReadFile(CompletePath("../testdata", "experiment.yaml"))
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
	os.Chdir(t.TempDir())
	exp := Experiment{
		Spec: ExperimentSpec{},
	}
	exp.initResults(1)

	exp.failExperiment()
	assert.False(t, exp.NoFailure())
}
