package base

import (
	"fmt"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	endpoint1 = "endpoint1"
	endpoint2 = "endpoint2"

	endpoint1URL = "https://something.com"
	endpoint2URL = "http://example.com"
)

func TestRunCollectHTTP(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	// // Exact URL match
	// httpmock.RegisterResponder("POST", "https://something.com",
	// 	httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			collectHTTPInputsHelper: collectHTTPInputsHelper{
				Duration:    StringPointer("1s"),
				PayloadFile: StringPointer(CompletePath("../", "testdata/payload/ukpolice.json")),
				Headers:     map[string]string{},
				URL:         "https://something.com",
			},
		},
	}

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err := ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)
}

func TestRunCollectHTTPSingleEndpoint(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)

	// httpmock.RegisterResponder("GET", endpoint1URL,
	// 	httpmock.NewStringResponder(200, ""))

	// httpmock.RegisterResponder("GET", endpoint2URL,
	// 	httpmock.NewStringResponder(200, ""))

	// httpmock.RegisterResponder("GET", "http://prometheus.istio-system:9090/api/v1/query",
	// 	func(req *http.Request) (*http.Response, error) {
	// 		if req.Header.Get(header1) == "" {
	// 			return httpmock.NewStringResponse(400, "Need header1"), nil
	// 		}

	// 		return httpmock.NewStringResponse(200, ""), nil
	// 	})

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			collectHTTPInputsHelper: collectHTTPInputsHelper{
				Duration: StringPointer("1s"),
			},
			Endpoints: map[string]collectHTTPInputsHelper{
				endpoint1: {
					URL: endpoint1URL,
				},
				endpoint2: {
					URL: endpoint2URL,
				},
			},
		},
	}

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err := ct.run(exp)
	assert.NoError(t, err)
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint1 + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	fmt.Println(exp.Result)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint1 + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint2 + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint2 + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)
}
