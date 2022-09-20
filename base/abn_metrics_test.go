package base

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockABNClient implements a mocked call to gRPC service
type mockABNClient struct {
	response []byte
}

func (c *mockABNClient) callGetApplicationJSON(appName string) (string, error) {
	return string(c.response), nil
}

func TestABNMetricsTask(t *testing.T) {
	fname := CompletePath("../testdata", "abninputs/application.yaml")
	b, err := os.ReadFile(filepath.Clean(fname))
	assert.NoError(t, err)

	task := &collectABNMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectABNMetrics),
		},
		With: ABNMetricsInputs{
			Application: "default/application",
		},
		client: &mockABNClient{
			response: b,
		},
	}
	exp := &Experiment{
		Spec:   []Task{task},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)

	err = task.run(exp)
	assert.NoError(t, err)

	// index 0: track = candidate; version = v2 expect count 252
	c := exp.Result.Insights.getSummaryAggregation(0, "abn/sample_metric", "count")
	assert.Equal(t, float64(252), *c)
	// index 1: track = default; version = v1; expect count 223
	c = exp.Result.Insights.getSummaryAggregation(1, "abn/sample_metric", "count")
	assert.Equal(t, float64(223), *c)
}
