package base

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	testNotifyURL = "https://test-service.com"
	templatePath  = "/template"
)

func getNotifyTask(t *testing.T, n notifyInputs) *notifyTask {
	// valid collect database task... should succeed
	nt := &notifyTask{
		TaskMeta: TaskMeta{
			Task: StringPointer(NotifyTaskName),
		},
		With: n,
	}

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip)
	return nt
}

// GET method
func TestNotify(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		Url:         testNotifyURL,
		SoftFailure: false,
	})

	// notify endpoint
	httpmock.RegisterResponder("GET", testNotifyURL,
		httpmock.NewStringResponder(200, "success"))

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should not fail
	assert.NoError(t, err)
}

type testNotification struct {
	Text       string `json:"text" yaml:"text"`
	TextReport string `json:"textReport" yaml:"textReport"`
	Report     Report `json:"report" yaml:"report"`
}

// POST method and PayloadTemplateURL
func TestNotifyWithPayload(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		Method:             "POST",
		Url:                testNotifyURL,
		PayloadTemplateURL: testNotifyURL + templatePath,
		SoftFailure:        false,
	})

	// payload template endpoint
	httpmock.RegisterResponder("GET", testNotifyURL+templatePath,
		httpmock.NewStringResponder(200, `{
	"text": "hello world",
	"textReport": "{{ regexReplaceAll "\"" (regexReplaceAll "\n" (.Report | toPrettyJson) "\\n") "\\\""}}",
	"report": {{ .Report | toPrettyJson }}
}`))

	// notify endpoint
	httpmock.RegisterResponder(
		"POST",
		testNotifyURL,
		func(req *http.Request) (*http.Response, error) {
			buf := new(bytes.Buffer)

			buf.ReadFrom(req.Body)

			// reqString := buf.String()
			// fmt.Println(reqString)

			var notification testNotification

			err := json.Unmarshal(buf.Bytes(), &notification)
			if err != nil {
				assert.Fail(t, "could not JSON unmarshal notification")
			}

			// check text
			assert.Equal(t, notification.Text, "hello world")

			// check textReport
			var textReportReport Report
			err = json.Unmarshal([]byte(notification.TextReport), &textReportReport)
			if err != nil {
				assert.Fail(t, "could not JSON unmarshal textReport in notification")
			}
			assert.Equal(t, textReportReport.NumTasks, 1)

			// check report
			assert.Equal(t, notification.Report.NumTasks, 1)

			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should not fail
	assert.NoError(t, err)
}

// GET method and headers and query parameters
func TestNotifyWithHeadersAndQueryParams(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		Url: testNotifyURL,
		Headers: map[string]string{
			"Hello": "headers",
		},
		Params: map[string]string{
			"hello": "params",
		},
		SoftFailure: false,
	})

	// notify endpoint
	httpmock.RegisterResponder(
		"GET",
		testNotifyURL,
		func(req *http.Request) (*http.Response, error) {
			// check headers
			assert.Equal(t, req.Header["Hello"], []string{"headers"})

			// check params
			assert.Equal(t, req.URL.Query()["hello"], []string{"params"})

			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should not fail
	assert.NoError(t, err)
}

// bad method and SoftFailure
func TestNotifyBadMethod(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		Url:         testNotifyURL,
		Method:      "abc",
		SoftFailure: false,
	})

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should fail
	assert.Error(t, err)

	nt = getNotifyTask(t, notifyInputs{
		Url:         testNotifyURL,
		Method:      "abc",
		SoftFailure: true,
	})

	exp = &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err = nt.run(exp)

	// test should not fail
	assert.NoError(t, err)
}

// default to POST method with PayloadTemplateURL
func TestNotifyPayloadTemplateURLDefaultMethod(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		Url:                testNotifyURL,
		PayloadTemplateURL: testNotifyURL + templatePath,
		SoftFailure:        false,
	})

	// payload template endpoint
	httpmock.RegisterResponder("GET", testNotifyURL+templatePath,
		httpmock.NewStringResponder(200, `hello world`))

	// notify endpoint
	httpmock.RegisterResponder(
		"GET",
		testNotifyURL,
		func(req *http.Request) (*http.Response, error) {
			assert.Fail(t, "notify task did not default to POST method with PayloadTemplateURL")

			return httpmock.NewStringResponse(400, "bad request"), nil
		},
	)

	// notify endpoint
	httpmock.RegisterResponder(
		"POST",
		testNotifyURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, "success"), nil
		},
	)

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should not fail
	assert.NoError(t, err)
}

// No URL
func TestNotifyNoURL(t *testing.T) {
	os.Chdir(t.TempDir())
	nt := getNotifyTask(t, notifyInputs{
		SoftFailure: false,
	})

	exp := &Experiment{
		Spec:   []Task{nt},
		Result: &ExperimentResult{},
	}
	exp.initResults(1)
	exp.Result.initInsightsWithNumVersions(1)

	err := nt.run(exp)

	// test should fail
	assert.Error(t, err)
}
