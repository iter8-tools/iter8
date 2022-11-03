package base

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	queryString           = "?query="
	exampleQueryParameter = "example query parameter\n"
	exampleRequestBody    = "example request body\n"

	// the provider URL is mocked
	cePromProviderURL  = "https://raw.githubusercontent.com/iter8-tools/iter8/master/testdata/metrics/test-ce.metrics.yaml"
	testCE             = "test-ce"
	testCEPromURL      = `test-database.com/prometheus/api/v1/query`
	testCERequestCount = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
	testCEErrorCount = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)\n"
	testCEErrorRate = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
	testCERequestCountWithRevisionName = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"
	testCEErrorCountWithRevisionName = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"
	testCEErrorRateWithRevisionName = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"

	// the provider URL is mocked
	testProviderURL = "https://raw.githubusercontent.com/iter8-tools/iter8/master/testdata/metrics/test-request-body.metrics.yaml"
	testRequestBody = "test-request-body"

	// the provider URL is mocked
	istioPromProviderURL = "https://raw.githubusercontent.com/iter8-tools/iter8/master/custommetrics/istio-prom.tpl"

	istioPromRequestCount = "sum(last_over_time(istio_requests_total{\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0)"
	istioPromErrorCount = "sum(last_over_time(istio_requests_total{\n" +
		"  response_code=~'5..',\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0)"
	istioPromErrorRate = "(sum(last_over_time(istio_requests_total{\n" +
		"  response_code=~'5..',\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0))/(sum(last_over_time(istio_requests_total{\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0))"
	istioPromMeanLatency = "(sum(last_over_time(istio_request_duration_milliseconds_sum{\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0))/(sum(last_over_time(istio_requests_total{\n" +
		"  destination_workload=\"myApp\",\n" +
		"  destination_workload_namespace=\"production\",\n" +
		"  reporter=\"destination\",\n" +
		"}[0s])) or on() vector(0))"
)

func getCustomMetricsTask(t *testing.T, providerName string, providerURL string) *customMetricsTask {
	// valid collect database task... should succeed
	ct := &customMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CustomMetricsTaskName),
		},
		With: customMetricsInputs{
			Templates: map[string]string{providerName: providerURL},
		},
	}

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
	return ct
}

// test getElapsedTimeSeconds()
func TestGetElapsedTimeSeconds(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	versionValues := map[string]interface{}{
		"startingTime": "2020-02-01T09:44:40Z",
	}

	exp := &Experiment{
		Spec:   []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionValues
	exp.initResults(1)

	elapsedTimeSeconds, _ := getElapsedTimeSeconds(versionValues, exp)

	// elapsedTimeSeconds should be a large number
	//
	// if getElapsedTimeSeconds() used the starting time from the experiment instead of
	// the one from versionValues, the elapsed time would be 0 or close to 0
	assert.Equal(t, elapsedTimeSeconds > 1000000, true)
}

// test if a user sets startingTime incorrectly getElapsedTimeSeconds()
func TestStartingTimeFormatError(t *testing.T) {
	_ = os.Chdir(t.TempDir())
	versionValues := map[string]interface{}{
		"startingTime": "1652935205",
	}

	exp := &Experiment{
		Spec:   []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionValues
	exp.initResults(1)
	_, err := getElapsedTimeSeconds(versionValues, exp)
	assert.Error(t, err)
}

// test istio-prom provider spec
func TestIstioProm(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "istio-prom.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, "istio-prom", istioPromProviderURL)
	ct.With.VersionValues = []map[string]interface{}{{
		"labels": map[string]interface{}{
			"reporter":                       "destination",
			"destination_workload":           "myApp",
			"destination_workload_namespace": "production",
		},
		"elapsedTimeSeconds": "5",
	}}

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// mock Istio Prometheus server
	httpmock.RegisterResponder("GET", "http://prometheus.istio-system:9090/api/v1/query",
		func(req *http.Request) (*http.Response, error) {
			queryParam := strings.TrimSpace(req.URL.Query().Get("query"))

			switch queryParam {
			case istioPromRequestCount:
				return httpmock.NewStringResponse(200, `{
					"status": "success",
					"data": {
						"resultType": "vector",
						"result": [
							{
								"metric": {},
								"value": [
									1645602108.839,
									"43"
								]
							}
						]
					}
				}`), nil

			case istioPromErrorCount:
				return httpmock.NewStringResponse(200, `{
						"status": "success",
						"data": {
							"resultType": "vector",
							"result": [
								{
									"metric": {},
									"value": [
										1645602108.839,
										"6"
									]
								}
							]
						}
					}`), nil

			case istioPromErrorRate:
				return httpmock.NewStringResponse(200, `{
						"status": "success",
						"data": {
							"resultType": "vector",
							"result": [
								{
									"metric": {},
									"value": [
										1645602108.839,
										"0.13953488372093023"
									]
								}
							]
						}
					}`), nil

			case istioPromMeanLatency:
				return httpmock.NewStringResponse(200, `{
						"status": "success",
						"data": {
							"resultType": "vector",
							"result": [
								{
									"metric": {},
									"value": [
										1645602108.839,
										"52"
									]
								}
							]
						}
					}`), nil

			}

			return nil, errors.New("")
		})

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// all three metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["istio-prom/request-count"][0], float64(43))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["istio-prom/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["istio-prom/error-rate"][0], 0.13953488372093023)
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["istio-prom/latency-mean"][0], float64(52))
}

// basic test with one version, mimicking Code Engine
// one version, three successful metrics
func TestCEOneVersion(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCount),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645602108.839,
								"43"
							]
						}
					]
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCount),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645648760.725,
								"6"
							]
						}
					]
				}
			}`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRate),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645043851.825,
								"0.13953488372093023"
							]
						}
					]
				}
			}`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// all three metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"][0], float64(43))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
}

// basic test with versionValues, mimicking Code Engine
// one version, three successful metrics
func TestCEVersionValues(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	ct.With.VersionValues = []map[string]interface{}{{
		"ibm_codeengine_revision_name": "v1",
	}}

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCountWithRevisionName),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645602108.839,
								"43"
							]
						}
					]
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCountWithRevisionName),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645648760.725,
								"6"
							]
						}
					]
				}
			}`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRateWithRevisionName),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645043851.825,
								"0.13953488372093023"
							]
						}
					]
				}
			}`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// all three metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"][0], float64(43))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
}

// test with one version and improper authorization, mimicking Code Engine
// one version, three successful metrics
func TestCEUnauthorized(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCount),
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCount),
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRate),
		httpmock.NewStringResponder(401, `Unauthorized`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// no values should be collected because of unauthorized requests
	assert.Equal(t, len(exp.Result.Insights.NonHistMetricValues[0]), 0)
}

// test with one version with some values, mimicking Code Engine
// one version, three successful metrics, one without values
func TestCESomeValues(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCount), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCount),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645648760.725,
								"6"
							]
						}
					]
				}
			}`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRate),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645043851.825,
								"0.13953488372093023"
							]
						}
					]
				}
			}`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersions(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	ct.With.VersionValues = []map[string]interface{}{{}, {}}

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCount), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCount),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645648760.725,
								"6"
							]
						}
					]
				}
			}`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRate),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645043851.825,
								"0.13953488372093023"
							]
						}
					]
				}
			}`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(2)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[1]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[1]["test-ce/error-rate"][0], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersionsAndMetrics(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", "test-ce.tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testCE, cePromProviderURL)

	ct.With.VersionValues = []map[string]interface{}{{}, {}}

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCERequestCount), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorCount),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645648760.725,
								"6"
							]
						}
					]
				}
			}`))

	// error-rate
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(testCEErrorRate),
		httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {},
							"value": [
								1645043851.825,
								"0.13953488372093023"
							]
						}
					]
				}
			}`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(2)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[1]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[1]["test-ce/error-rate"][0], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)
}

// basic test with a request body
func TestRequestBody(t *testing.T) {
	dat, err := os.ReadFile(CompletePath("../testdata/custommetrics", testRequestBody+".tpl"))
	assert.NoError(t, err)
	tplString := string(dat)

	_ = os.Chdir(t.TempDir())
	ct := getCustomMetricsTask(t, testRequestBody, testProviderURL)

	// mock provider URL
	httpmock.RegisterResponder("GET", istioPromProviderURL,
		httpmock.NewStringResponder(200, tplString))

	// request-count
	httpmock.RegisterResponder("GET", testCEPromURL+queryString+url.QueryEscape(exampleQueryParameter),
		func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				if err != nil {
					panic(err)
				}

				if string(b) == exampleRequestBody {
					return httpmock.NewStringResponse(200, `{
							"status": "success",
							"data": {
								"resultType": "vector",
								"result": [
									{
										"metric": {},
										"value": [
											1645602108.839,
											"43"
										]
									}
								]
							}
						}`), nil
				}
			}

			return nil, errors.New("")
		})

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	_ = exp.Result.initInsightsWithNumVersions(1)

	err = ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0][testRequestBody+"/request-count"][0], float64(43))
}
