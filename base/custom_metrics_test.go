package base

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	testCEURL          = "https://raw.githubusercontent.com/iter8-tools/iter8/master/testdata/metrics/test-ce.metrics.yaml"
	testRequestBodyURL = "https://raw.githubusercontent.com/iter8-tools/iter8/master/testdata/metrics/test-request-body.metrics.yaml"
	testRequestBody    = "test-request-body"
	testPromURL        = `test-database.com/prometheus/api/v1/query`
	queryString        = "?query="
	requestCountQuery  = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
	errorCountQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)\n"
	errorRateQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
	requestCountWithRevisionNameQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"
	errorCountWithRevisionNameQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"
	errorRateWithRevisionNameQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_revision_name=\"v1\",\n" +
		"}[0s])) or on() vector(0)\n"
	exampleQueryParameter = "example query parameter\n"
	exampleRequestBody    = "example request body\n"
)

// test getElapsedTimeSeconds()
func TestGetElapsedTimeSeconds(t *testing.T) {
	os.Chdir(t.TempDir())
	versionInfo := map[string]interface{}{
		"startingTime": "2020-02-01T09:44:40Z",
	}

	exp := &Experiment{
		Spec:   []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionInfo
	exp.initResults(1)

	elapsedTimeSeconds, _ := getElapsedTimeSeconds(versionInfo, exp)

	// elapsedTimeSeconds should be a large number
	//
	// if getElapsedTimeSeconds() used the starting time from the experiment instead of
	// the one from versionInfo, the elapsed time would be 0 or close to 0
	assert.Equal(t, elapsedTimeSeconds > 1000000, true)
}

// test if a user sets startingTime incorrectly getElapsedTimeSeconds()
func TestStartingTimeFormatError(t *testing.T) {
	os.Chdir(t.TempDir())
	versionInfo := map[string]interface{}{
		"startingTime": "1652935205",
	}

	exp := &Experiment{
		Spec:   []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionInfo
	exp.initResults(1)
	_, err := getElapsedTimeSeconds(versionInfo, exp)
	assert.Error(t, err)
}

// some reusable test code
func headForTests(t *testing.T, providerURL string) *customMetricsTask {
	// values := map[string]interface{}{
	// 	"providerURL": "http://prometheus.istio-system:9090/api/v1/query",
	// 	"IAMToken":    "test-token",
	// 	"GUID":        "test-guid",
	// }

	// valid collect database task... should succeed
	ct := &customMetricsTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CustomMetricsTaskName),
		},
		With: customMetricsInputs{
			ProviderURLs: []string{providerURL},
			// Values:       values,
		},
	}

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
	return ct
}

// basic test with one version, mimicking Code Engine
// one version, three successful metrics
func TestCEOneVersion(t *testing.T) {
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateQuery),
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// all three metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"][0], float64(43))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

}

// basic test with versionInfo, mimicking Code Engine
// one version, three successful metrics
func TestCEVersionInfo(t *testing.T) {
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)
	ct.With.VersionValues = []map[string]interface{}{{
		"ibm_codeengine_revision_name": "v1",
	}}

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountWithRevisionNameQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountWithRevisionNameQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateWithRevisionNameQuery),
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := ct.run(exp)

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
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountQuery),
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountQuery),
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-rate
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateQuery),
		httpmock.NewStringResponder(401, `Unauthorized`))

	exp := &Experiment{
		Spec:   []Task{ct},
		Result: &ExperimentResult{},
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	// no values should be collected because of unauthorized requests
	assert.Equal(t, len(exp.Result.Insights.NonHistMetricValues[0]), 0)

}

// test with one version with some values, mimicking Code Engine
// one version, three successful metrics, one without values
func TestCESomeValues(t *testing.T) {
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateQuery),
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := ct.run(exp)

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
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)

	ct.With.VersionValues = []map[string]interface{}{{}, {}}

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateQuery),
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(2)

	err := ct.run(exp)

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
	os.Chdir(t.TempDir())
	ct := headForTests(t, testCEURL)
	ct.With.VersionValues = []map[string]interface{}{{}, {}}

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

	// error-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorCountQuery),
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
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(errorRateQuery),
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(2)

	err := ct.run(exp)

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
	os.Chdir(t.TempDir())
	ct := headForTests(t, testRequestBodyURL)

	// request-count
	httpmock.RegisterResponder("GET", testPromURL+queryString+url.QueryEscape(exampleQueryParameter),
		func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				b, err := ioutil.ReadAll(req.Body)
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
		driver: &mockDriver{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := ct.run(exp)

	// test should not fail
	assert.NoError(t, err)

	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0][testRequestBody+"/request-count"][0], float64(43))
}
