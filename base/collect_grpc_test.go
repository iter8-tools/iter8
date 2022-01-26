package base

import (
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/iter8-tools/iter8/base/internal"
	"github.com/iter8-tools/iter8/base/internal/helloworld/helloworld"
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
		taskMeta: taskMeta{
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
