package base

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"fortio.org/fortio/fhttp"
	"github.com/stretchr/testify/assert"
)

const (
	endpoint1 = "endpoint1"
	endpoint2 = "endpoint2"

	foo  = "foo"
	bar  = "bar"
	from = "from"
)

func TestRunCollectHTTP(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)

	// /foo/ handler
	called := false // ensure that the /foo/ handler is called
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		data, _ := io.ReadAll(r.Body)
		testData, _ := os.ReadFile(CompletePath("../", "testdata/payload/ukpolice.json"))

		// assert that PayloadFile is working
		assert.True(t, bytes.Equal(data, testData))

		w.WriteHeader(200)
	}
	mux.HandleFunc("/"+foo, handler)

	baseURL := fmt.Sprintf("http://localhost:%d/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration:    StringPointer("1s"),
				PayloadFile: StringPointer(CompletePath("../", "testdata/payload/ukpolice.json")),
				Headers:     map[string]string{},
				URL:         baseURL + foo,
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
	assert.True(t, called) // ensure that the /foo/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

	mm, err = exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "/" + builtInHTTPLatencyPercentilePrefix + "50")
	assert.NotNil(t, mm)
	assert.NoError(t, err)
}

// Multiple endpoints are provided
// Test both the /foo/ and /bar/ endpoints
// Test both endpoints have their respective header values
func TestRunCollectHTTPMultipleEndpoints(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)

	// /foo/ handler
	fooCalled := false // ensure that the /foo/ handler is called
	fooHandler := func(w http.ResponseWriter, r *http.Request) {
		fooCalled = true

		// assert "from" header has value "foo"
		assert.Equal(t, foo, r.Header.Get(from))

		w.WriteHeader(200)
	}
	mux.HandleFunc("/"+foo, fooHandler)

	// /bar/ handler
	barCalled := false // ensure that the /foo/ handler is called
	barHandler := func(w http.ResponseWriter, r *http.Request) {
		barCalled = true

		// assert "from" header has value "bar"
		assert.Equal(t, bar, r.Header.Get(from))

		w.WriteHeader(200)
	}
	mux.HandleFunc("/"+bar, barHandler)

	baseURL := fmt.Sprintf("http://localhost:%d/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration: StringPointer("1s"),
			},
			Endpoints: map[string]endpoint{
				endpoint1: {
					URL: baseURL + foo,
					Headers: map[string]string{
						from: foo,
					},
				},
				endpoint2: {
					URL: baseURL + bar,
					Headers: map[string]string{
						from: bar,
					},
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
	assert.True(t, fooCalled) // ensure that the /foo/ handler is called
	assert.True(t, barCalled) // ensure that the /bar/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint1 + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

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

// Multiple endpoints are provided but they share one URL
// Test that the base-level URL is provided to each endpoint
// Make multiple calls to the same URL but with different headers
func TestRunCollectHTTPSingleEndpointMultipleCalls(t *testing.T) {
	mux, addr := fhttp.DynamicHTTPServer(false)

	// handler
	fooCalled := false // ensure that foo header is provided
	barCalled := false // ensure that bar header is provided
	fooHandler := func(w http.ResponseWriter, r *http.Request) {
		from := r.Header.Get(from)
		if from == foo {
			fooCalled = true
		} else if from == bar {
			barCalled = true
		}

		w.WriteHeader(200)
	}
	mux.HandleFunc("/", fooHandler)

	baseURL := fmt.Sprintf("http://localhost:%d/", addr.Port)

	// valid collect HTTP task... should succeed
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration: StringPointer("1s"),
				URL:      baseURL,
			},
			Endpoints: map[string]endpoint{
				endpoint1: {
					Headers: map[string]string{
						from: foo,
					},
				},
				endpoint2: {
					Headers: map[string]string{
						from: bar,
					},
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
	assert.True(t, fooCalled) // ensure that the /foo/ handler is called
	assert.True(t, barCalled) // ensure that the /bar/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	mm, err := exp.Result.Insights.GetMetricsInfo(httpMetricPrefix + "-" + endpoint1 + "/" + builtInHTTPLatencyMeanID)
	assert.NotNil(t, mm)
	assert.NoError(t, err)

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

func TestErrorCode(t *testing.T) {
	task := collectHTTPTask{}
	assert.True(t, task.errorCode(-1))

	// if no lower limit (check upper)
	upper := 10
	task.With.ErrorRanges = append(task.With.ErrorRanges, errorRange{
		Upper: &upper,
	})
	assert.True(t, task.errorCode(5))

	// if no upper limit (check lower)
	task.With.ErrorRanges = []errorRange{}
	lower := 1
	task.With.ErrorRanges = append(task.With.ErrorRanges, errorRange{
		Lower: &lower,
	})
	assert.True(t, task.errorCode(5))

	// if both limits are present (check both)
	task.With.ErrorRanges = []errorRange{}
	task.With.ErrorRanges = append(task.With.ErrorRanges, errorRange{
		Upper: &upper,
		Lower: &lower,
	})
	assert.True(t, task.errorCode(5))
}
