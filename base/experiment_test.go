package base

import (
	"fmt"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
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
	assert.Equal(t, 1, len(e.Spec))

	b, err = os.ReadFile(CompletePath("../testdata", "experiment_grpc.yaml"))
	assert.NoError(t, err)
	e = &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.Spec))

	b, err = os.ReadFile(CompletePath("../testdata", "experiment_db.yaml"))
	assert.NoError(t, err)
	e = &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.Spec))
}

func TestRunningTasks(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", GetTrackingHandler(&verifyHandlerCalled))

	// valid collect task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration: StringPointer("1s"),
				Headers:  map[string]string{},
				URL:      url,
			},
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
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
}

func TestRunExperiment(t *testing.T) {
	_ = os.Chdir(t.TempDir())

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", GetTrackingHandler(&verifyHandlerCalled))

	// create experiment.yaml
	CreateExperimentYaml(t, CompletePath("../testdata", "experiment.tpl"), url, "experiment.yaml")
	b, err := os.ReadFile("experiment.yaml")

	assert.NoError(t, err)
	e := &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.Spec))

	err = RunExperiment(false, &mockDriver{e})
	assert.NoError(t, err)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)

	assert.True(t, e.Completed())
	assert.True(t, e.NoFailure())
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
