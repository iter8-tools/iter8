package base

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bojand/ghz/runner"
	"github.com/iter8-tools/iter8/base/internal"
	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/stretchr/testify/assert"
)

const (
	unary         = "unary"
	unary2        = "unary2"
	server        = "server"
	client        = "client"
	bidirectional = "bidirectional"
)

// Credit: Several of the tests in this file are based on
// https://github.com/bojand/ghz/blob/master/runner/run_test.go
func TestRunCollectGRPCUnary(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	call := "helloworld.Greeter.SayHello"

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
				Call: call,
				Host: internal.LocalHostPort,
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

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

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	taskData := exp.Result.Insights.TaskData[CollectGRPCTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	ghzResult := GHZResult{}
	err = json.Unmarshal(taskDataBytes, &ghzResult)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(ghzResult))
	assert.NotNil(t, ghzResult[call])
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
func TestRunCollectGRPCMultipleEndpoints(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)
	err = ct.run(exp)

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	taskData := exp.Result.Insights.TaskData[CollectGRPCTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	ghzResult := GHZResult{}
	err = json.Unmarshal(taskDataBytes, &ghzResult)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(ghzResult))
	assert.NotNil(t, ghzResult[unary])
	assert.NotNil(t, ghzResult[server])
	assert.NotNil(t, ghzResult[client])
	assert.NotNil(t, ghzResult[bidirectional])
}

// TODO: should this still return insights even though the endpoints cannot be reached?
// This would mean no Grafana dashboard would be produced
//
// If the endpoints cannot be reached, then do not throw an error
// Should not return an nil pointer dereference error (see #1451)
func TestRunCollectGRPCMultipleNoEndpoints(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	unaryCall := "helloworld.Greeter.SayHello"
	serverCall := "helloworld.Greeter.SayHelloCS"
	clientCall := "helloworld.Greeter.SayHellos"
	bidirectionalCall := "helloworld.Greeter.SayHelloBidi"

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
					Call: unaryCall,
				},
				server: {
					Data: map[string]interface{}{"name": "bob"},
					Call: serverCall,
				},
				client: {
					Data: map[string]interface{}{"name": "bob"},
					Call: clientCall,
				},
				bidirectional: {
					Data: map[string]interface{}{"name": "bob"},
					Call: bidirectionalCall,
				},
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

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

	taskData := exp.Result.Insights.TaskData[CollectGRPCTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	ghzResult := GHZResult{}
	err = json.Unmarshal(taskDataBytes, &ghzResult)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(ghzResult))
}

func TestRunCollectGRPCSingleEndpointMultipleCalls(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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
				Call: "helloworld.Greeter.SayHello",
			},
			Endpoints: map[string]runner.Config{
				unary: {
					Data: map[string]interface{}{"name": "bob"},
				},
				unary2: {
					Data: map[string]interface{}{"name": "charles"},
				},
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

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

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	count := gs.GetCount(callType)
	assert.Equal(t, 400, count)

	taskData := exp.Result.Insights.TaskData[CollectGRPCTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	ghzResult := GHZResult{}
	err = json.Unmarshal(taskDataBytes, &ghzResult)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(ghzResult))
	assert.NotNil(t, ghzResult[unary])
	assert.NotNil(t, ghzResult[unary2])
}

func TestRunCollectGRPCWithWarmup(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	call := "helloworld.Greeter.SayHello"

	_ = os.Chdir(t.TempDir())
	callType := helloworld.Unary
	gs, s, err := internal.StartServer(false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	t.Cleanup(s.Stop)

	// valid collect GRPC task... should succeed
	warmupTrue := true
	ct := &collectGRPCTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectGRPCTaskName),
		},
		With: collectGRPCInputs{
			Config: runner.Config{
				Data: map[string]interface{}{"name": "bob"},
				Call: call,
				Host: internal.LocalHostPort,
			},
			Warmup: &warmupTrue,
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

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

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	assert.NoError(t, err)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	// warmup option ensures that ghz results are not written to insights
	assert.Nil(t, exp.Result.Insights)
}

// Credit: Several of the tests in this file are based on
// https://github.com/bojand/ghz/blob/master/runner/run_test.go
func TestRunCollectGRPCWithIncorrectNumVersions(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	call := "helloworld.Greeter.SayHello"

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
				Call: call,
				Host: internal.LocalHostPort,
			},
		},
	}

	log.Logger.Debug("dial timeout before defaulting... ", ct.With.DialTimeout.String())

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)

	exp.Result.Insights = &Insights{
		NumVersions: 2, // will cause grpc task to fail; grpc task expects insights been nil or numVersions set to 1
	}

	err = ct.run(exp)

	log.Logger.Debug("dial timeout after defaulting... ", ct.With.DialTimeout.String())

	// fail because of initInsightsWithNumVersions()
	assert.Error(t, err)

	count := gs.GetCount(callType)
	assert.Equal(t, 200, count)

	// error ensures that ghz results are not written to insights
	assert.Nil(t, exp.Result.Insights.TaskData)
}
