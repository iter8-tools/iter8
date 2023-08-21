package base

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// If the endpoint does not exist, fail gracefully
// Should not return an nil pointer dereference error (see #1451)
func TestRunCollectGRPCUnaryNoEndpoint(t *testing.T) {
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
	err := ct.run(exp)

	// Error should be a connection error, not a nil pointer dereference error
	// Test written like this because of conversion between localhost and 127.0.0.1
	assert.True(t, strings.HasPrefix(err.Error(), "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp"))
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

// If the endpoints cannot be reached, then do not throw an error
// Should not return an nil pointer dereference error (see #1451)
func TestRunCollectGRPCMultipleNoEndpoints(t *testing.T) {
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
	err := ct.run(exp)
	assert.NoError(t, err)

	// No metrics should be collected
	assert.Equal(t, 0, len(exp.Result.Insights.NonHistMetricValues[0]))
	assert.Equal(t, 0, len(exp.Result.Insights.HistMetricValues[0]))
	assert.Equal(t, 0, len(exp.Result.Insights.SummaryMetricValues[0]))
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

	exp := &Experiment{
		Spec: []Task{ct},
	}

	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err = exp.Spec[0].run(exp)
	assert.NoError(t, err)

	expjson, _ := json.Marshal(exp)
	fmt.Println(string(expjson))

	expBytes, _ := yaml.Marshal(exp)
	log.Logger.Debug("\n" + string(expBytes))

	count := gs.GetCount(callType)
	assert.Equal(t, int(ct.With.N), count)
}
