package base

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"testing"
	"text/template"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	metricsDirectory  = "../testdata/metrics/"
	testCe            = "test-ce"
	testPromURL       = `test-database.com/prometheus/api/v1/query?query=`
	requestCountQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
	errorCountQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)\n"
	errorRateQuery = "sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"  ibm_codeengine_status!=\"200\",\n" +
		"}[0s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{\n" +
		"}[0s])) or on() vector(0)\n"
)

type collectDatabaseTemplateInput struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	IAMToken string `json:"IAMToken" yaml:"IAMToken"`
	GUID     string `json:"GUID" yaml:"GUID"`
}

/*
	The collect database task checks for metric files in the current directory.

	The test metrics files are stored in ../testdata/ so in order for the tests
	to run, the metrics files are copied to a temprorary directory and the current
	working directory is changed to the temporary directory
*/
func GoToTempDirectoryAndCopyMetricsFile(t *testing.T, test func()) error {
	metricsFileName := testCe + experimentMetricsPathSuffix

	originalPath, err := os.Getwd()
	if err != nil {
		return err
	}

	// return to original path
	t.Cleanup(func() {
		os.Chdir(originalPath)
	})

	// get metrics file
	srcFile, err := os.Open(metricsDirectory + metricsFileName)
	if err != nil {
		return errors.New("could not open metrics file.")
	}
	t.Cleanup(func() {
		srcFile.Close()
	})

	// go to temp directory
	destDir := t.TempDir()
	os.Chdir(destDir)

	// create copy of metrics file in temp directory
	destFile, err := os.Create(metricsFileName)
	if err != nil {
		return errors.New("could not create copy of metrics file in temp directory.")
	}
	t.Cleanup(func() {
		destFile.Close()
	})
	io.Copy(destFile, srcFile)

	// run test
	test()

	return nil
}

// test getElapsedTimeSeconds()
func TestGetElapsedTimeSeconds(t *testing.T) {
	versionInfo := map[string]interface{}{
		"ibm_service_instance": "version1",
		"startingTime":         "Feb 4, 2014 at 6:05pm (PST)",
	}

	exp := &Experiment{
		Tasks:  []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionInfo
	exp.initResults()

	elapsedTimeSeconds, _ := getElapsedTimeSeconds(versionInfo, exp)

	// elapsedTimeSeconds should be a large number
	//
	// if getElapsedTimeSeconds() used the starting time from the experiment instead of
	// the one from versionInfo, the elapsed time would be 0 or close to 0
	assert.Equal(t, elapsedTimeSeconds > 1000000, true)
}

// basic test with one version, mimicking Code Engine
// one version, three successful metrics
func TestCEOneVersion(t *testing.T) {
	err := GoToTempDirectoryAndCopyMetricsFile(t, func() {
		input := &collectDatabaseTemplateInput{
			Endpoint: "test-database.com",
			IAMToken: "test-token",
			GUID:     "test-guid",
		}

		// convert input to map[string]interface{}
		var templateInput map[string]interface{}
		inrec, err := json.Marshal(input)
		assert.NoError(t, err)

		json.Unmarshal(inrec, &templateInput)

		// valid collect database task... should succeed
		ct := &collectDatabaseTask{
			TaskMeta: TaskMeta{
				Task: StringPointer(CollectDatabaseTaskName),
			},
			With: collectDatabaseInputs{
				Providers: []string{testCe},
				VersionInfo: []map[string]interface{}{{
					"ibm_service_instance": "version1",
				}},
			},
		}

		httpmock.Activate()

		// request-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(requestCountQuery),
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
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorCountQuery),
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
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorRateQuery),
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

		template, err := template.ParseFiles(testCe + experimentMetricsPathSuffix)

		assert.NoError(t, err)

		md := mockDriver{
			metricsTemplate: template,
		}

		exp := &Experiment{
			Tasks:  []Task{ct},
			Result: &ExperimentResult{},
			driver: &md,
		}
		exp.initResults()
		exp.Result.initInsightsWithNumVersions(1)

		err = ct.run(exp)

		// test should not fail
		assert.NoError(t, err)

		// all three metrics should exist and have values
		assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"][0], float64(43))
		assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
		assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

		httpmock.DeactivateAndReset()
	})

	assert.NoError(t, err)
}

// test with one version and improper authorization, mimicking Code Engine
// one version, three successful metrics
func TestCEUnauthorized(t *testing.T) {
	err := GoToTempDirectoryAndCopyMetricsFile(t, func() {
		input := &collectDatabaseTemplateInput{
			Endpoint: "test-database.com",
			IAMToken: "test-token",
			GUID:     "test-guid",
		}

		// convert input to map[string]interface{}
		var templateInput map[string]interface{}
		inrec, err := json.Marshal(input)
		assert.NoError(t, err)

		json.Unmarshal(inrec, &templateInput)

		ct := &collectDatabaseTask{
			TaskMeta: TaskMeta{
				Task: StringPointer(CollectDatabaseTaskName),
			},
			With: collectDatabaseInputs{
				Providers: []string{testCe},
				VersionInfo: []map[string]interface{}{{
					"ibm_service_instance": "version1",
				}},
			},
		}

		httpmock.Activate()

		// request-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(requestCountQuery),
			httpmock.NewStringResponder(401, `Unauthorized`))

		// error-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorCountQuery),
			httpmock.NewStringResponder(401, `Unauthorized`))

		// error-rate
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorRateQuery),
			httpmock.NewStringResponder(401, `Unauthorized`))

		template, err := template.ParseFiles(testCe + experimentMetricsPathSuffix)

		assert.NoError(t, err)

		md := mockDriver{
			metricsTemplate: template,
		}

		exp := &Experiment{
			Tasks:  []Task{ct},
			Result: &ExperimentResult{},
			driver: &md,
		}
		exp.initResults()
		exp.Result.initInsightsWithNumVersions(1)

		err = ct.run(exp)

		// test should not fail
		assert.NoError(t, err)

		// no values should be collected because of unauthorized requests
		assert.Equal(t, len(exp.Result.Insights.NonHistMetricValues[0]), 0)

		httpmock.DeactivateAndReset()
	})

	assert.NoError(t, err)
}

// test with one version with some values, mimicking Code Engine
// one version, three successful metrics, one without values
func TestCESomeValues(t *testing.T) {
	err := GoToTempDirectoryAndCopyMetricsFile(t, func() {
		input := &collectDatabaseTemplateInput{
			Endpoint: "test-database.com",
			IAMToken: "test-token",
			GUID:     "test-guid",
		}

		// convert input to map[string]interface{}
		var templateInput map[string]interface{}
		inrec, err := json.Marshal(input)
		assert.NoError(t, err)

		json.Unmarshal(inrec, &templateInput)

		ct := &collectDatabaseTask{
			TaskMeta: TaskMeta{
				Task: StringPointer(CollectDatabaseTaskName),
			},
			With: collectDatabaseInputs{
				Providers: []string{testCe},
				VersionInfo: []map[string]interface{}{{
					"ibm_service_instance": "version1",
				}},
			},
		}

		httpmock.Activate()

		// request-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

		// error-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorCountQuery),
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
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorRateQuery),
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

		template, err := template.ParseFiles(testCe + experimentMetricsPathSuffix)

		assert.NoError(t, err)

		md := mockDriver{
			metricsTemplate: template,
		}

		exp := &Experiment{
			Tasks:  []Task{ct},
			Result: &ExperimentResult{},
			driver: &md,
		}
		exp.initResults()
		exp.Result.initInsightsWithNumVersions(1)

		err = ct.run(exp)

		// test should not fail
		assert.NoError(t, err)

		// two metrics should exist and have values
		assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
		assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

		// request-count should not exist because there was no value from response
		_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
		assert.Equal(t, ok, false)

		httpmock.DeactivateAndReset()
	})

	assert.NoError(t, err)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersions(t *testing.T) {
	err := GoToTempDirectoryAndCopyMetricsFile(t, func() {
		input := &collectDatabaseTemplateInput{
			Endpoint: "test-database.com",
			IAMToken: "test-token",
			GUID:     "test-guid",
		}

		// convert input to map[string]interface{}
		var templateInput map[string]interface{}
		inrec, err := json.Marshal(input)
		assert.NoError(t, err)

		json.Unmarshal(inrec, &templateInput)

		ct := &collectDatabaseTask{
			TaskMeta: TaskMeta{
				Task: StringPointer(CollectDatabaseTaskName),
			},
			With: collectDatabaseInputs{
				Providers: []string{testCe},
				VersionInfo: []map[string]interface{}{{
					"ibm_service_instance": "version1",
				}, {
					"ibm_service_instance": "version2",
				}},
			},
		}

		httpmock.Activate()

		// request-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

		// error-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorCountQuery),
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
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorRateQuery),
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

		template, err := template.ParseFiles(testCe + experimentMetricsPathSuffix)

		assert.NoError(t, err)

		md := mockDriver{
			metricsTemplate: template,
		}

		exp := &Experiment{
			Tasks:  []Task{ct},
			Result: &ExperimentResult{},
			driver: &md,
		}
		exp.initResults()
		exp.Result.initInsightsWithNumVersions(2)

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

		httpmock.DeactivateAndReset()
	})

	assert.NoError(t, err)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersionsAndMetrics(t *testing.T) {
	err := GoToTempDirectoryAndCopyMetricsFile(t, func() {
		input := &collectDatabaseTemplateInput{
			Endpoint: "test-database.com",
			IAMToken: "test-token",
			GUID:     "test-guid",
		}

		// convert input to map[string]interface{}
		var templateInput map[string]interface{}
		inrec, err := json.Marshal(input)
		assert.NoError(t, err)

		json.Unmarshal(inrec, &templateInput)

		ct := &collectDatabaseTask{
			TaskMeta: TaskMeta{
				Task: StringPointer(CollectDatabaseTaskName),
			},
			With: collectDatabaseInputs{
				Providers: []string{testCe},
				VersionInfo: []map[string]interface{}{{
					"ibm_service_instance": "version1",
				}, {
					"ibm_service_instance": "version2",
				}},
			},
		}

		httpmock.Activate()

		// request-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(requestCountQuery), httpmock.NewStringResponder(200, `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": []
				}
			}`))

		// error-count
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorCountQuery),
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
		httpmock.RegisterResponder("GET", testPromURL+url.QueryEscape(errorRateQuery),
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

		template, err := template.ParseFiles(testCe + experimentMetricsPathSuffix)

		assert.NoError(t, err)

		md := mockDriver{
			metricsTemplate: template,
		}

		exp := &Experiment{
			Tasks:  []Task{ct},
			Result: &ExperimentResult{},
			driver: &md,
		}
		exp.initResults()
		exp.Result.initInsightsWithNumVersions(2)

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

		httpmock.DeactivateAndReset()
	})

	assert.NoError(t, err)
}
