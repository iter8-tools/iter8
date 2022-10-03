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
			Task: StringPointer(CollectABNMetricsTaskName),
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

	assertCount(t, exp.Result.Insights, "default", float64(223))
	assertCount(t, exp.Result.Insights, "candidate", float64(252))
}

func assertCount(t *testing.T, in *Insights, track string, count float64) bool {
	for i, vn := range in.VersionNames {
		if vn.Track == track {
			c := in.getSummaryAggregation(i, "abn/sample_metric", "count")
			return assert.Equal(t, count, *c)
		}
	}
	return false
}
