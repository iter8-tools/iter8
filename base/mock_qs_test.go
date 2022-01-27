package base

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestMockQuickStartWithSLOs(t *testing.T) {
	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration: StringPointer("1s"),
			VersionInfo: []*versionHTTP{{
				Headers: map[string]string{},
				URL:     "https://example.com",
			}},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     "built-in/http-latency-mean",
				UpperLimit: float64Pointer(100),
			}},
		},
	}
	exp := &Experiment{
		Tasks: []Task{ct, at},
	}

	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)
	err := exp.Tasks[0].Run(exp)
	assert.NoError(t, err)
	err = exp.Tasks[1].Run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied {
		for _, b := range v {
			assert.True(t, b)
		}
	}
}

func TestMockQuickStartWithSLOsAndPercentiles(t *testing.T) {
	// mock the http endpoint
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://example.com",
		httpmock.NewStringResponder(200, `all good`))

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			Duration: StringPointer("1s"),
			VersionInfo: []*versionHTTP{{
				Headers: map[string]string{},
				URL:     "https://example.com",
			}},
		},
	}

	at := &assessTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(AssessTaskName),
		},
		With: assessInputs{
			SLOs: []SLO{{
				Metric:     "built-in/http-latency-mean",
				UpperLimit: float64Pointer(100),
			}, {
				Metric:     "built-in/http-latency-p95.00",
				UpperLimit: float64Pointer(200),
			}},
		},
	}
	exp := &Experiment{
		Tasks: []Task{ct, at},
	}

	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)
	err := exp.Tasks[0].Run(exp)
	assert.NoError(t, err)
	err = exp.Tasks[1].Run(exp)
	assert.NoError(t, err)
	// assert SLOs are satisfied
	for _, v := range exp.Result.Insights.SLOsSatisfied {
		for _, b := range v {
			assert.True(t, b)
		}
	}
}