package base

import (
	"fmt"
	"testing"
	"time"

	"fortio.org/fortio/fhttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestRunCollectHTTP(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)
	mux.HandleFunc("/echo1/", fhttp.EchoHandler)
	testURL := fmt.Sprintf("http://localhost:%d/echo1/", addr.Port)

	httpmock.Activate()
	defer httpmock.Deactivate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://data.police.uk/api/crimes-street-dates",
		httpmock.NewStringResponder(200, `[{"my": 1, "great": "payload"}]`))

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			PayloadURL:  StringPointer("https://data.police.uk/api/crimes-street-dates"),
			VersionInfo: []*fhttp.HTTPRunnerOptions{},
		},
	}
	ct.With.Duration, _ = time.ParseDuration("1s")
	ct.With.VersionInfo[0].URL = testURL

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	err := ct.Run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(iter8BuiltInPrefix + "/" + builtInHTTPLatencyMeanId)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(iter8BuiltInPrefix + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)

}
