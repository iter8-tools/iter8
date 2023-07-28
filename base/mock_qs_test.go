package base

import (
	"fmt"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/stretchr/testify/assert"
)

func TestMockQuickStartWithSLOs(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration: StringPointer("2s"),
				Headers:  map[string]string{},
				URL:      testURL,
			},
		},
	}

	exp := &Experiment{
		Spec: []Task{ct},
	}

	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err := exp.Spec[0].run(exp)
	assert.NoError(t, err)
}

func TestMockQuickStartWithSLOsAndPercentiles(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration: StringPointer("1s"),
				Headers:  map[string]string{},
				URL:      testURL,
			},
		},
	}

	exp := &Experiment{
		Spec: []Task{ct},
	}

	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err := exp.Spec[0].run(exp)
	assert.NoError(t, err)
}
