package base

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

func TestRunningTasks(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", GetTrackingHandler(&verifyHandlerCalled))

	_ = os.Chdir(t.TempDir())

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
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)
	err = ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
	// sanity check -- handler was called
	assert.True(t, verifyHandlerCalled)
}

func TestRunExperiment(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	// create and configure HTTP endpoint for testing
	mux, addr := fhttp.DynamicHTTPServer(false)
	url := fmt.Sprintf("http://127.0.0.1:%d/get", addr.Port)
	var verifyHandlerCalled bool
	mux.HandleFunc("/get", GetTrackingHandler(&verifyHandlerCalled))

	// mock metrics server
	StartHTTPMock(t)
	metricsServerCalled := false
	MockMetricsServer(MockMetricsServerInput{
		MetricsServerURL: metricsServerURL,
		ExperimentResultCallback: func(req *http.Request) {
			metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("test"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyExperimentResult := ExperimentResult{}
			err = json.Unmarshal(body, &bodyExperimentResult)
			assert.NoError(t, err)
			assert.NotNil(t, body)
		},
	})

	_ = os.Chdir(t.TempDir())

	// create experiment.yaml
	CreateExperimentYaml(t, CompletePath("../testdata", "experiment.tpl"), url, "experiment.yaml")
	b, err := os.ReadFile("experiment.yaml")

	assert.NoError(t, err)
	e := &Experiment{}
	err = yaml.Unmarshal(b, e)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.Spec))

	err = RunExperiment(&mockDriver{e})
	assert.NoError(t, err)
	assert.True(t, metricsServerCalled)
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

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		specBytes  string
		errMessage string
	}{
		{
			specBytes: `[{"task":"ready"}]`,
		},
		{
			specBytes: `[{"task":"http"}]`,
		},
		{
			specBytes: `[{"task":"grpc"}]`,
		},
		{
			specBytes: `[{"task":"notify"}]`,
		},
	}

	for _, test := range tests {
		exp := ExperimentSpec{}
		err := exp.UnmarshalJSON([]byte(test.specBytes))
		assert.NoError(t, err)
	}
}

func TestUnmarshalJSONError(t *testing.T) {
	tests := []struct {
		specBytes  string
		errMessage string
	}{
		{
			specBytes:  "hello world",
			errMessage: `invalid character 'h' looking for beginning of value`,
		},
		{
			specBytes:  "[{}]",
			errMessage: `invalid task found without a task name or a run command`,
		},
		{
			specBytes:  `[{"task":"hello world"}]`,
			errMessage: `unknown task: hello world`,
		},
	}

	for _, test := range tests {
		exp := ExperimentSpec{}
		err := exp.UnmarshalJSON([]byte(test.specBytes))
		assert.Error(t, err)
		assert.EqualError(t, err, test.errMessage)
	}
}

func TestInitInsightsWithNumVersions(t *testing.T) {
	r := ExperimentResult{
		Insights: &Insights{
			NumVersions: 1,
		},
	}

	err := r.initInsightsWithNumVersions(1)
	assert.NoError(t, err)

	// Mismatching version numbers
	err = r.initInsightsWithNumVersions(2)
	assert.Error(t, err)
}
