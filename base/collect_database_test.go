package base

import (
	"os"
	"testing"
	"text/template"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var templatePath string = "../testdata/templates/ce.metrics.tpl"
var tempMetricsPath string = "test-ce.metrics.yaml"

// has to be a map[string]string in order to do input checks in template
// func executeTemplate(inputs interface{}, templatePath string, writePath string) error {
func executeTemplate(inputs map[string]interface{}, templatePath string, writePath string) error {
	template, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	file, err := os.Create(writePath)
	if err != nil {
		return err
	}

	err = template.Execute(file, inputs)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// test getElapsedTime()
func TestGetElapsedTime(t *testing.T) {
	versionInfo := map[string]interface{}{
		"ibm_service_instance": "version1",
		"StartingTime":         1000,
	}

	exp := &Experiment{
		Tasks:  []Task{},
		Result: &ExperimentResult{},
	}

	// this should add a startingTime that will be overwritten by the one in
	// versionInfo
	exp.InitResults()

	elapsedTime, _ := getElapsedTime(versionInfo, exp)

	// elapsedTime should be a large number
	//
	// if getElapsedTime() used the starting time from the experiment instead of
	// the one from versionInfo, the elapsed time would be 0 or close to 0
	assert.Equal(t, elapsedTime > 1000000, true)
}

// basic test with one version, mimicking Code Engine
// one version, three successful metrics
func TestCEOneVersion(t *testing.T) {
	// create metrics file from template
	inputs := map[string]interface{}{
		"MonitoringEndpoint": "test-database.com",
		"IAMToken":           "test-token",
		"GUID":               "test-guid",
	}
	err := executeTemplate(inputs, templatePath, tempMetricsPath)

	assert.NoError(t, err)

	// valid collect database task... should succeed
	ct := &collectDatabaseTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectDatabaseTaskName),
		},
		With: collectDatabaseInputs{
			VersionInfo: []map[string]interface{}{{
				"ibm_service_instance": "version1",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// request-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
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
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29++%0A`,
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
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29%2Fsum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
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
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)

	err = ct.Run(exp)

	// test should not fail
	assert.NoError(t, err)

	// all three metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"][0], float64(43))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

	// delete metrics file
	os.Remove(tempMetricsPath)
}

// test with one version and improper authorization, mimicking Code Engine
// one version, three successful metrics
func TestCEUnauthorized(t *testing.T) {
	// create metrics file from template
	inputs := map[string]interface{}{
		"MonitoringEndpoint": "test-database.com",
		"IAMToken":           "test-token",
		"GUID":               "test-guid",
	}
	err := executeTemplate(inputs, templatePath, tempMetricsPath)

	assert.NoError(t, err)

	ct := &collectDatabaseTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectDatabaseTaskName),
		},
		With: collectDatabaseInputs{
			VersionInfo: []map[string]interface{}{{
				"ibm_service_instance": "version1",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// request-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29++%0A`,
		httpmock.NewStringResponder(401, `Unauthorized`))

	// error-rate
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29%2Fsum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
		httpmock.NewStringResponder(401, `Unauthorized`))

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)

	err = ct.Run(exp)

	// test should not fail
	assert.NoError(t, err)

	// no values should be collected because of unauthorized requests
	assert.Equal(t, len(exp.Result.Insights.NonHistMetricValues[0]), 0)

	// delete metrics file
	os.Remove(tempMetricsPath)
}

// test with one version with some values, mimicking Code Engine
// one version, three successful metrics, one without values
func TestCESomeValues(t *testing.T) {
	// create metrics file from template
	inputs := map[string]interface{}{
		"MonitoringEndpoint": "test-database.com",
		"IAMToken":           "test-token",
		"GUID":               "test-guid",
	}
	err := executeTemplate(inputs, templatePath, tempMetricsPath)

	assert.NoError(t, err)

	ct := &collectDatabaseTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectDatabaseTaskName),
		},
		With: collectDatabaseInputs{
			VersionInfo: []map[string]interface{}{{
				"ibm_service_instance": "version1",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// request-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
		httpmock.NewStringResponder(200, `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": []
			}
		}`))

	// error-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29++%0A`,
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
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29%2Fsum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
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
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)

	err = ct.Run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)

	// delete metrics file
	os.Remove(tempMetricsPath)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersions(t *testing.T) {
	// create metrics file from template
	inputs := map[string]interface{}{
		"MonitoringEndpoint": "test-database.com",
		"IAMToken":           "test-token",
		"GUID":               "test-guid",
	}
	err := executeTemplate(inputs, templatePath, tempMetricsPath)

	assert.NoError(t, err)

	ct := &collectDatabaseTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectDatabaseTaskName),
		},
		With: collectDatabaseInputs{
			VersionInfo: []map[string]interface{}{{
				"ibm_service_instance": "version1",
			}, {
				"ibm_service_instance": "version2",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// request-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
		httpmock.NewStringResponder(200, `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": []
			}
		}`))

	// error-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29++%0A`,
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
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29%2Fsum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
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
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(1)

	err = ct.Run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][1], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][1], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)

	// delete metrics file
	os.Remove(tempMetricsPath)
}

// test with two version with some values, mimicking Code Engine
// two versions, four successful metrics, two without values
func TestCEMultipleVersionsAndMetrics(t *testing.T) {
	// create metrics file from template
	inputs := map[string]interface{}{
		"MonitoringEndpoint": "test-database.com",
		"IAMToken":           "test-token",
		"GUID":               "test-guid",
	}
	err := executeTemplate(inputs, templatePath, tempMetricsPath)

	assert.NoError(t, err)

	ct := &collectDatabaseTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectDatabaseTaskName),
		},
		With: collectDatabaseInputs{
			VersionInfo: []map[string]interface{}{{
				"ibm_service_instance": "version1",
			}, {
				"ibm_service_instance": "version2",
			}},
		},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// request-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
		httpmock.NewStringResponder(200, `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": []
			}
		}`))

	// error-count
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29++%0A`,
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
	httpmock.RegisterResponder("GET", `test-database.com/prometheus/api/v1/query?query=sum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A++ibm_codeengine_status%21%3D%22200%22%2C%0A%7D%5B0s%5D%29%29%2Fsum%28last_over_time%28ibm_codeengine_application_requests_total%7B%0A%7D%5B0s%5D%29%29+%0A`,
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
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	exp.Result.initInsightsWithNumVersions(2)

	err = ct.Run(exp)

	// test should not fail
	assert.NoError(t, err)

	// two metrics should exist and have values
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][0], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-count"][1], float64(6))
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][0], 0.13953488372093023)
	assert.Equal(t, exp.Result.Insights.NonHistMetricValues[0]["test-ce/error-rate"][1], 0.13953488372093023)

	// request-count should not exist because there was no value from response
	_, ok := exp.Result.Insights.NonHistMetricValues[0]["test-ce/request-count"]
	assert.Equal(t, ok, false)

	// delete metrics file
	os.Remove(tempMetricsPath)
}
