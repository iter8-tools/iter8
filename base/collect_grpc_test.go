package base

import (
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/iter8-tools/iter8/base/internal"
	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
)

// Credit: Several of the tests in this file are based on
// https://github.com/bojand/ghz/blob/master/runner/run_test.go
func TestRunCollectGRPCUnary(t *testing.T) {
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer s.Stop()

	// valid collect GRPC task... should succeed
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGPRCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				N:           1,
				C:           1,
				Timeout:     runner.Duration(20 * time.Second),
				Data:        map[string]interface{}{"name": "bob"},
				DialTimeout: runner.Duration(20 * time.Second),
			},
			ProtoURL: StringPointer("https://raw.githubusercontent.com/bojand/ghz/v0.105.0/testdata/greeter.proto"),
			VersionInfo: []*versionGRPC{{
				Call: "helloworld.Greeter.SayHello",
				Host: internal.TestLocalhost,
			}},
		},
	}

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err = ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 1, count)
}

func TestMockGRPCWithSLOsAndPercentiles(t *testing.T) {
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer s.Stop()

	// valid collect GRPC task... should succeed
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGPRCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				N:           100,
				RPS:         20,
				C:           1,
				Timeout:     runner.Duration(20 * time.Second),
				Data:        map[string]interface{}{"name": "bob"},
				DialTimeout: runner.Duration(20 * time.Second),
			},
			ProtoURL: StringPointer("https://raw.githubusercontent.com/bojand/ghz/v0.105.0/testdata/greeter.proto"),
			VersionInfo: []*versionGRPC{{
				Call: "helloworld.Greeter.SayHello",
				Host: internal.TestLocalhost,
			}},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     "built-in/grpc-latency/mean",
				UpperLimit: float64Pointer(100),
			}, {
				Metric:     "built-in/grpc-latency/p95.00",
				UpperLimit: float64Pointer(200),
			}, {
				Metric:     "built-in/grpc-error-count",
				UpperLimit: float64Pointer(0),
			}},
		},
	}
	exp := &Experiment{
		Tasks: []Task{ct, at},
	}

	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)
	err = exp.Tasks[0].Run(exp)
	assert.NoError(t, err)
	err = exp.Tasks[1].Run(exp)
	assert.NoError(t, err)

	// print metrics
	log.Logger.Info("Metric info", exp.Result.Insights.MetricsInfo)
	log.Logger.Info("Hist metric values", exp.Result.Insights.HistMetricValues)
	log.Logger.Info("Non hist metric values", exp.Result.Insights.NonHistMetricValues)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied {
		for _, b := range v {
			assert.True(t, b)
		}
	}

	count := gs.GetCount(callType)
	assert.Equal(t, int(ct.With.N), count)

}
