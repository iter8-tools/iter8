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
			collectHTTPInputsHelper: collectHTTPInputsHelper{
				Duration: StringPointer("2s"),
				Headers:  map[string]string{},
				URL:      testURL,
			},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: &SLOLimits{
				Upper: []SLO{{
					Metric: "http/latency-mean",
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
	err := exp.Spec[0].run(exp)
	assert.NoError(t, err)
	err = exp.Spec[1].run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied.Upper {
		for _, b := range v {
			assert.True(t, b)
		}
	}
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
			collectHTTPInputsHelper: collectHTTPInputsHelper{
				Duration: StringPointer("1s"),
				Headers:  map[string]string{},
				URL:      testURL,
			},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: &SLOLimits{
				Upper: []SLO{{
					Metric: "http/latency-mean",
					Limit:  100,
				}, {
					Metric: "http/latency-p95.00",
					Limit:  200,
				}},
			},
		},
	}
	exp := &Experiment{
		Spec: []Task{ct, at},
	}

	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)
	err := exp.Spec[0].run(exp)
	assert.NoError(t, err)
	err = exp.Spec[1].run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied.Upper {
		for _, b := range v {
			assert.True(t, b)
		}
	}
}
