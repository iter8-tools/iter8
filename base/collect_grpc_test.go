package base

import (
	"os"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/iter8-tools/iter8/base/internal"
	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

const (
	unary         = "unary"
	server        = "server"
	client        = "client"
	bidirectional = "bidirectional"
)

// Credit: Several of the tests in this file are based on
// https://github.com/bojand/ghz/blob/master/runner/run_test.go
func TestRunCollectGRPCUnary(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	t.Cleanup(s.Stop)

	// valid collect GRPC task... should succeed
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGRPCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				Data: map[string]interface{}{"name": "bob"},
				Call: "helloworld.Greeter.SayHello",
				Host: internal.LocalHostPort,
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err = ct.run(exp)

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	mm, err := exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "/" + gRPCErrorCountMetricName)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "/" + gRPCLatencySampleMetricName)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "/" + gRPCLatencySampleMetricName + "/" + string(MaxAggregator))
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "/" + gRPCLatencySampleMetricName + "/" + PercentileAggregatorPrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)
}

// Credit: Several of the tests in this file are based on
// https://github.com/bojand/ghz/blob/master/runner/run_test.go
func TestRunCollectGRPCEndpoints(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	t.Cleanup(s.Stop)

	// valid collect GRPC task... should succeed
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGRPCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				Host: internal.LocalHostPort,
			},
			Endpoints: map[string]runner.Config{
				unary: {
					Data: map[string]interface{}{"name": "bob"},
					Call: "helloworld.Greeter.SayHello",
				},
				server: {
					Data: map[string]interface{}{"name": "bob"},
					Call: "helloworld.Greeter.SayHelloCS",
				},
				client: {
					Data: map[string]interface{}{"name": "bob"},
					Call: "helloworld.Greeter.SayHellos",
				},
				bidirectional: {
					Data: map[string]interface{}{"name": "bob"},
					Call: "helloworld.Greeter.SayHelloBidi",
				},
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err = ct.run(exp)

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	grpcMethods := []string{unary, server, client, bidirectional}
	for _, method := range grpcMethods {
		mm, err := exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "-" + method + "/" + gRPCErrorCountMetricName)
		assert.NotNil(t, mm)
		assert.NoError(t, err)

		mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "-" + method + "/" + gRPCLatencySampleMetricName)
		assert.NotNil(t, mm)
		assert.NoError(t, err)

		mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "-" + method + "/" + gRPCLatencySampleMetricName + "/" + string(MaxAggregator))
		assert.NotNil(t, mm)
		assert.NoError(t, err)

		mm, err = exp.Result.Insights.GetMetricsInfo(gRPCMetricPrefix + "-" + method + "/" + gRPCLatencySampleMetricName + "/" + PercentileAggregatorPrefix + "50")
		assert.NotNil(t, mm)
		assert.NoError(t, err)
	}
}

func TestMockGRPCWithSLOsAndPercentiles(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	t.Cleanup(s.Stop)

	// valid collect GRPC task... should succeed
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGRPCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				N:           100,
				RPS:         20,
				C:           1,
				Timeout:     runner.Duration(20 * time.Second),
				Data:        map[string]interface{}{"name": "bob"},
				DialTimeout: runner.Duration(20 * time.Second),
				Call:        "helloworld.Greeter.SayHello",
				Host:        internal.LocalHostPort,
			},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: &SLOLimits{
				Lower: []SLO{{
					Metric: "grpc/request-count",
					Limit:  100,
				}},
				Upper: []SLO{{
					Metric: "grpc/latency/mean",
					Limit:  100,
				}, {
					Metric: "grpc/latency/p95.00",
					Limit:  200,
				}, {
					Metric: "grpc/latency/stddev",
					Limit:  20,
				}, {
					Metric: "grpc/latency/max",
					Limit:  200,
				}, {
					Metric: "grpc/error-count",
					Limit:  0,
				}, {
					Metric: "grpc/request-count",
					Limit:  100,
				}},
			},
		},
	}
	exp := &Experiment{
		Spec: []Task{ct, at},
	}

	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err = exp.Spec[0].run(exp)
	assert.NoError(t, err)
	err = exp.Spec[1].run(exp)
	assert.NoError(t, err)

	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied.Upper {
		for _, b := range v {
			assert.True(t, b)
		}
	}
	for _, v := range exp.Result.Insights.SLOsSatisfied.Lower {
		for _, b := range v {
			assert.True(t, b)
		}
	}

	expBytes, _ := yaml.Marshal(exp)
	log.Logger.Debug("\n" + string(expBytes))

	count := gs.GetCount(callType)
	assert.Equal(t, int(ct.With.N), count)
}
