package base

import (
	"fmt"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/stretchr/testify/assert"
)

func TestMockQuickStartWithSLOs(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration: StringPointer("2s"),
			Headers:  map[string]string{},
			URL:      testURL,
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     "http/latency-mean",
				UpperLimit: float64Pointer(100),
			}},
		},
	}
	exp := &Experiment{
		Tasks: []Task{ct, at},
	}

	exp.initResults()
	exp.Result.initInsightsWithNumVersions(1)
	err := exp.Tasks[0].run(exp)
	assert.NoError(t, err)
	err = exp.Tasks[1].run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied {
		for _, b := range v {
			assert.True(t, b)
		}
	}
}

func TestMockQuickStartWithSLOsAndPercentiles(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration: StringPointer("1s"),
			Headers:  map[string]string{},
			URL:      testURL,
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     "http/latency-mean",
				UpperLimit: float64Pointer(100),
			}, {
				Metric:     "http/latency-p95.00",
				UpperLimit: float64Pointer(200),
			}},
		},
	}
	exp := &Experiment{
		Tasks: []Task{ct, at},
	}

	exp.initResults()
	exp.Result.initInsightsWithNumVersions(1)
	err := exp.Tasks[0].run(exp)
	assert.NoError(t, err)
	err = exp.Tasks[1].run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied {
		for _, b := range v {
			assert.True(t, b)
		}
	}
}
