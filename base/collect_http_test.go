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
	"github.com/jarcoal/httpmock"
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

func startHTTPMock(t *testing.T) {
	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
}

type DashboardCallback func(req *http.Request)

type mockMetricsServerInput struct {
	metricsServerURL string

	// GET /httpDashboard
	httpDashboardCallback DashboardCallback
	// GET /grpcDashboard
	gRPCDashboardCallback DashboardCallback
	// PUT /performanceResult
	performanceResultCallback DashboardCallback
}

func mockMetricsServer(input mockMetricsServerInput) {
	// GET /httpDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.metricsServerURL+HTTPDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.httpDashboardCallback != nil {
				input.httpDashboardCallback(req)
			}

			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// GET /grpcDashboard
	httpmock.RegisterResponder(
		http.MethodGet,
		input.metricsServerURL+GRPCDashboardPath,
		func(req *http.Request) (*http.Response, error) {
			if input.gRPCDashboardCallback != nil {
				input.gRPCDashboardCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	// PUT /performanceResult
	httpmock.RegisterResponder(
		http.MethodPut,
		input.metricsServerURL+PerformanceResultPath,
		func(req *http.Request) (*http.Response, error) {
			if input.performanceResultCallback != nil {
				input.performanceResultCallback(req)
			}
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)
}

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

	// mock metrics server
	startHTTPMock(t)
	metricsServerCalled := false
	mockMetricsServer(mockMetricsServerInput{
		metricsServerURL: metricsServerURL,
		performanceResultCallback: func(req *http.Request) {
			metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("experiment"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyFortioResult := FortioResult{}
			err = json.Unmarshal(body, &bodyFortioResult)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			if _, ok := bodyFortioResult.EndpointResults[url]; !ok {
				assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain url: %s", url))
			}
		},
	})

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
	assert.True(t, metricsServerCalled) // ensure that the metrics server is called
	assert.True(t, called)              // ensure that the /foo/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
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
	endpoint1URL := baseURL + foo
	endpoint2URL := baseURL + bar

	// mock metrics server
	startHTTPMock(t)
	metricsServerCalled := false
	mockMetricsServer(mockMetricsServerInput{
		metricsServerURL: metricsServerURL,
		performanceResultCallback: func(req *http.Request) {
			metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("experiment"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyFortioResult := FortioResult{}
			err = json.Unmarshal(body, &bodyFortioResult)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			if _, ok := bodyFortioResult.EndpointResults[endpoint1URL]; !ok {
				assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain url: %s", endpoint1URL))
			}

			if _, ok := bodyFortioResult.EndpointResults[endpoint2URL]; !ok {
				assert.Fail(t, fmt.Sprintf("payload FortioResult.EndpointResult does not contain url: %s", endpoint2URL))
			}
		},
	})

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
	assert.True(t, metricsServerCalled) // ensure that the metrics server is called
	assert.True(t, fooCalled)           // ensure that the /foo/ handler is called
	assert.True(t, barCalled)           // ensure that the /bar/ handler is called
	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
}

// TODO: this test is broken because the FortioResult.EndpointResults uses URL
// as the key but in this case, there are two endpoints with the same URL but
// different headers.
//
// // Multiple endpoints are provided but they share one URL
// // Test that the base-level URL is provided to each endpoint
// // Make multiple calls to the same URL but with different headers
// func TestRunCollectHTTPSingleEndpointMultipleCalls(t *testing.T) {
// 	mux, addr := fhttp.DynamicHTTPServer(false)

// 	// handler
// 	fooCalled := false // ensure that foo header is provided
// 	barCalled := false // ensure that bar header is provided
// 	fooHandler := func(w http.ResponseWriter, r *http.Request) {
// 		from := r.Header.Get(from)
// 		if from == foo {
// 			fooCalled = true
// 		} else if from == bar {
// 			barCalled = true
// 		}

// 		w.WriteHeader(200)
// 	}
// 	mux.HandleFunc("/", fooHandler)

// 	baseURL := fmt.Sprintf("http://localhost:%d/", addr.Port)

// 	// valid collect HTTP task... should succeed
// 	ct := &collectHTTPTask{
// 		TaskMeta: TaskMeta{
// 			Task: StringPointer(CollectHTTPTaskName),
// 		},
// 		With: collectHTTPInputs{
// 			endpoint: endpoint{
// 				Duration: StringPointer("1s"),
// 				URL:      baseURL,
// 			},
// 			Endpoints: map[string]endpoint{
// 				endpoint1: {
// 					Headers: map[string]string{
// 						from: foo,
// 					},
// 				},
// 				endpoint2: {
// 					Headers: map[string]string{
// 						from: bar,
// 					},
// 				},
// 			},
// 		},
// 	}

// 	exp := &Experiment{
// 		Spec:   []Task{ct},
// 		Result: &ExperimentResult{},
// 	}
// 	exp.initResults(1)
// 	err := ct.run(exp)
// 	assert.NoError(t, err)
// 	assert.True(t, fooCalled) // ensure that the /foo/ handler is called
// 	assert.True(t, barCalled) // ensure that the /bar/ handler is called
// 	assert.Equal(t, exp.Result.Insights.NumVersions, 1)
// }

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

	// mock metrics server
	startHTTPMock(t)
	// metricsServerCalled := false
	mockMetricsServer(mockMetricsServerInput{
		metricsServerURL: metricsServerURL,
		performanceResultCallback: func(req *http.Request) {
			// metricsServerCalled = true

			// check query parameters
			assert.Equal(t, myName, req.URL.Query().Get("experiment"))
			assert.Equal(t, myNamespace, req.URL.Query().Get("namespace"))

			// check payload
			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.NotNil(t, body)

			// check payload content
			bodyFortioResult := FortioResult{}
			err = json.Unmarshal(body, &bodyFortioResult)
			assert.NoError(t, err)

			// no EndpointResults because endpoints cannot be reached
			assert.Equal(t, `{"EndpointResults":{},"Summary":{"numVersions":1,"versionNames":null}}`, string(body))
		},
	})

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

func TestPutPerformanceResultToMetricsService(t *testing.T) {
	startHTTPMock(t)

	metricsServerURL := "http://my-server.com"
	namespace := "my-namespace"
	experiment := "my-experiment"
	data := map[string]string{
		"hello": "world",
	}

	called := false
	httpmock.RegisterResponder(http.MethodPut, metricsServerURL+PerformanceResultPath,
		func(req *http.Request) (*http.Response, error) {
			called = true

			assert.Equal(t, namespace, req.URL.Query().Get("namespace"))
			assert.Equal(t, experiment, req.URL.Query().Get("experiment"))

			body, err := io.ReadAll(req.Body)
			assert.NoError(t, err)
			assert.Equal(t, "{\"hello\":\"world\"}", string(body))

			return httpmock.NewStringResponse(200, "success"), nil
		})

	err := putPerformanceResultToMetricsService(
		metricsServerURL,
		namespace,
		experiment,
		data,
	)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRunCollectHTTPGrafana(t *testing.T) {
	// METRICS_SERVER_URL must be provided
	metricsServerURL := "http://iter8.default:8080"
	err := os.Setenv("METRICS_SERVER_URL", metricsServerURL)
	assert.NoError(t, err)

	// mock metrics server
	metricsServerCalled := false
	namespace := "default"
	experiment := "default"
	startHTTPMock(t)
	httpmock.RegisterResponder(http.MethodPut, metricsServerURL+PerformanceResultPath,
		func(req *http.Request) (*http.Response, error) {
			metricsServerCalled = true

			assert.Equal(t, namespace, req.URL.Query().Get("namespace"))
			assert.Equal(t, experiment, req.URL.Query().Get("experiment"))

			return httpmock.NewStringResponse(200, "success"), nil
		})

	mux, addr := fhttp.DynamicHTTPServer(false)

	// mock endpoint
	endpointCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		endpointCalled = true

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
				URL: baseURL + foo,
			},
		},
	}

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
		Metadata: ExperimentMetadata{
			Namespace: "default",
			Name:      "default",
		},
	}
	exp.initResults(1)
	err = ct.run(exp)
	assert.NoError(t, err)
	assert.True(t, metricsServerCalled)
	assert.True(t, endpointCalled)
}
