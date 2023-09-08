package base

import (
	"bytes"
	"encoding/json"
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

	myName      = "myName"
	myNamespace = "myNamespace"
)

func TestRunCollectHTTP(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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

	url := fmt.Sprintf("http://localhost:%d/", addr.Port) + foo

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
				URL:         url,
			},
		},
	}

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
	assert.True(t, called) // ensure that the /foo/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	taskData := exp.Result.Insights.TaskData[CollectHTTPTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	httpResult := HTTPResult{}
	err = json.Unmarshal(taskDataBytes, &httpResult)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(httpResult))
	assert.NotNil(t, httpResult[url])
}

// If the endpoint does not exist, fail gracefully
// Should not return an nil pointer dereference error (see #1451)
func TestRunCollectHTTPNoEndpoint(t *testing.T) {
	_, addr := fhttp.DynamicHTTPServer(false)

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
				URL:         baseURL,
			},
		},
	}

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	err := ct.run(exp)

	assert.EqualError(t, err, fmt.Sprintf("error 404 for %s (176 bytes)", baseURL))
}

// Multiple endpoints are provided
// Test both the /foo/ and /bar/ endpoints
// Test both endpoints have their respective header values
func TestRunCollectHTTPMultipleEndpoints(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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
	endpoint1 := "endpoint1"
	endpoint2 := "endpoint2"
	endpoint1URL := baseURL + foo
	endpoint2URL := baseURL + bar

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
					URL: endpoint1URL,
					Headers: map[string]string{
						from: foo,
					},
				},
				endpoint2: {
					URL: endpoint2URL,
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
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)
	err = ct.run(exp)
	assert.NoError(t, err)
	assert.True(t, fooCalled) // ensure that the /foo/ handler is called
	assert.True(t, barCalled) // ensure that the /bar/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	taskData := exp.Result.Insights.TaskData[CollectHTTPTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	httpResult := HTTPResult{}
	err = json.Unmarshal(taskDataBytes, &httpResult)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(httpResult))
	assert.NotNil(t, httpResult[endpoint1])
	assert.NotNil(t, httpResult[endpoint2])
}

// Multiple endpoints are provided but they share one URL
// Test that the base-level URL is provided to each endpoint
// Make multiple calls to the same URL but with different headers
func TestRunCollectHTTPSingleEndpointMultipleCalls(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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
	endpoint1 := "endpoint1"
	endpoint2 := "endpoint2"

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
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)
	err = ct.run(exp)
	assert.NoError(t, err)
	assert.True(t, fooCalled) // ensure that the /foo/ handler is called
	assert.True(t, barCalled) // ensure that the /bar/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)

	taskData := exp.Result.Insights.TaskData[CollectHTTPTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	httpResult := HTTPResult{}
	err = json.Unmarshal(taskDataBytes, &httpResult)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(httpResult))
	assert.NotNil(t, httpResult[endpoint1])
	assert.NotNil(t, httpResult[endpoint2])
}

// TODO: should this still return insights even though the endpoints cannot be reached?
// This would mean no Grafana dashboard would be produced
//
// If the endpoints cannot be reached, then do not throw an error
// Should not return an nil pointer dereference error (see #1451)
func TestRunCollectHTTPMultipleNoEndpoints(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

	_, addr := fhttp.DynamicHTTPServer(false)

	baseURL := fmt.Sprintf("http://localhost:%d/", addr.Port)
	endpoint1URL := baseURL + foo
	endpoint2URL := baseURL + bar

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
					URL: endpoint1URL,
					Headers: map[string]string{
						from: foo,
					},
				},
				endpoint2: {
					URL: endpoint2URL,
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
		Metadata: ExperimentMetadata{
			Name:      myName,
			Namespace: myNamespace,
		},
	}
	exp.initResults(1)
	err = ct.run(exp)
	assert.NoError(t, err)

	taskData := exp.Result.Insights.TaskData[CollectHTTPTaskName]
	assert.NotNil(t, taskData)

	taskDataBytes, err := json.Marshal(taskData)
	assert.NoError(t, err)
	httpResult := HTTPResult{}
	err = json.Unmarshal(taskDataBytes, &httpResult)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(httpResult))
}

func TestRunCollectHTTPWithWarmup(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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

	url := fmt.Sprintf("http://localhost:%d/", addr.Port) + foo

	// valid collect HTTP task... should succeed
	warmupTrue := true
	ct := &collectHTTPTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectHTTPTaskName),
		},
		With: collectHTTPInputs{
			endpoint: endpoint{
				Duration:    StringPointer("1s"),
				PayloadFile: StringPointer(CompletePath("../", "testdata/payload/ukpolice.json")),
				Headers:     map[string]string{},
				URL:         url,
			},
			Warmup: &warmupTrue,
		},
	}

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
	assert.True(t, called) // ensure that the /foo/ handler is called

	// warmup option ensures that Fortio results are not written to insights
	assert.Nil(t, exp.Result.Insights)
}

func TestRunCollectHTTPWithIncorrectNumVersions(t *testing.T) {
	// define METRICS_SERVER_URL
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv(MetricsServerURL, metricsServerURL)
	assert.NoError(t, err)

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

	url := fmt.Sprintf("http://localhost:%d/", addr.Port) + foo

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
				URL:         url,
			},
		},
	}

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
		NumVersions: 2, // will cause http task to fail; grpc task expects insights been nil or numVersions set to 1
	}

	err = ct.run(exp)
	assert.Error(t, err) // fail because of initInsightsWithNumVersions()

	assert.True(t, called) // ensure that the /foo/ handler is called

	// error ensures that Fortio results are not written to insights
	assert.Nil(t, exp.Result.Insights.TaskData)
}

func TestGetFortioOptions(t *testing.T) {
	// check to catch nil QPS
	_, err := getFortioOptions(endpoint{})
	assert.Error(t, err)

	// check for catch nil connections
	QPS := float32(8)
	_, err = getFortioOptions(endpoint{
		QPS: &QPS,
	})
	assert.Error(t, err)

	// check to catch nil allowInitialErrors
	connections := 8
	_, err = getFortioOptions(endpoint{
		QPS:         &QPS,
		Connections: &connections,
	})
	assert.Error(t, err)

	numRequests := int64(5)
	contentType := "testType"
	payloadStr := "testPayload"
	allowInitialErrors := true

	options, err := getFortioOptions(endpoint{
		NumRequests:        &numRequests,
		ContentType:        &contentType,
		PayloadStr:         &payloadStr,
		QPS:                &QPS,
		Connections:        &connections,
		AllowInitialErrors: &allowInitialErrors,
	})

	assert.NoError(t, err)

	s, _ := json.Marshal(options)
	fmt.Println(string(s))

	assert.Equal(t, numRequests, options.RunnerOptions.Exactly)
	assert.Equal(t, contentType, options.ContentType)
	assert.Equal(t, []byte(payloadStr), options.Payload)
}
