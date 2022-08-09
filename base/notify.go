package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"github.com/iter8-tools/iter8/base/log"
)

// notifyInputs is the input to the notify task
type notifyInputs struct {
	// URL is the URL of the notification hook
	Url string `json:"url" yaml:"url"`

	// Method is the HTTP method that needs to be used
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Params is the set of HTTP parameters that need to be sent
	Params map[string]string `json:"params,omitempty" yaml:"params,omitempty"`

	// Headers is the set of HTTP headers that need to be sent
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// URL is the URL of the request payload template that should be used
	PayloadTemplateURL string `json:"payloadTemplateURL,omitempty" yaml:"payloadTemplateURL,omitempty"`

	// SoftFailure indicates the task and experiment should not fail if the task
	// cannot successfully send a request to the notification hook
	SoftFailure bool `json:"softFailure" yaml:"softFailure"`
}

const (
	// NotifyTaskName is the task name
	NotifyTaskName = "notify"
)

// notifyTask sends notifications
type notifyTask struct {
	TaskMeta
	With notifyInputs `json:"with" yaml:"with"`
}

// Report is the data that is given to the payload template
type Report struct {
	// Group string `json:"group" yaml:"group"`

	// SLOs map[string]SLOReport `json:"SLOs" yaml:"SLOs"`

	// Metrics map[string]MetricReport `json:"metrics" yaml:"metrics"`

	// Timestamp is when the report was created
	// For example: 2022-08-09 15:10:36.569745 -0400 EDT m=+12.599643189
	TimeStamp string `json:"timeStamp" yaml:"timeStamp"`

	// Completed is whether or not the experiment has completed
	Completed bool `json:"completed" yaml:"completed"`

	// NoTaskFailures is whether or not the experiment had any tasks that failed
	NoTaskFailures bool `json:"noTaskFailures" yaml:"noTaskFailures"`

	// NumTasks is the number of tasks in the experiment
	NumTasks int `json:"numTasks" yaml:"numTasks"`

	// NumCompletedTasks is the number of completed tasks in the experiment
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// NumLoops is the current loop of the experiment
	NumLoops int `json:"numLoops" yaml:"numLoops"`

	// Experiment is the experiment struct
	Experiment *Experiment `json:"experiment" yaml:"experiment"`
}

// NotifyPayloadTemplateValues contains the report as well as the report in some other formats
type NotifyPayloadTemplateValues struct {
	// Report is the report
	Report Report `json:"report" yaml:"report"`

	// JSONStringReport is the report but marshalled into JSON and stringified
	JSONStringReport string `json:"JSONStringReport" yaml:"JSONStringReport"`

	// EscapedJSONStringReport is the report but marshalled into JSON and stringified and with tabs and new lines escaped
	EscapedJSONStringReport string `json:"escapedJSONStringReport" yaml:"escapedJSONStringReport"`
}

// getPayloadTemplateValues gets the values for the payload tempalte
func getPayloadTemplateValues(exp *Experiment) (*NotifyPayloadTemplateValues, error) {
	report := Report{
		// Group: exp.driver.Group
		// SLOs:           slos,
		// Metrics:        metrics,

		TimeStamp:         time.Now().String(),
		Completed:         exp.Completed(),
		NoTaskFailures:    exp.NoFailure(),
		NumTasks:          len(exp.Spec),
		NumCompletedTasks: exp.Result.NumCompletedTasks,
		NumLoops:          exp.Result.NumLoops,
		Experiment:        exp,
	}

	marshalledReport, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Logger.Error("could not JSON marshall report")
		return nil, err
	}

	stringReport := string(marshalledReport)

	// escape double quotes, tabs, and new lines
	escapedReport := strings.Replace(stringReport, "\"", "\\\"", -1)
	escapedReport = strings.Replace(escapedReport, "\n", "\\n", -1)
	escapedReport = strings.Replace(escapedReport, "\t", "\\t", -1)

	return &NotifyPayloadTemplateValues{
		Report:                  report,
		JSONStringReport:        stringReport,
		EscapedJSONStringReport: escapedReport,
	}, nil
}

// getPayload fetches the payload template from the PayloadTemplateURL and
// executes it with values from getPayloadTemplateValues()
func (t *notifyTask) getPayload(exp *Experiment) (string, error) {
	if t.With.PayloadTemplateURL != "" {
		template, err := getProviderTemplate(t.With.PayloadTemplateURL)
		if err != nil {
			log.Logger.Error("could not get payload template")
			return "", err
		}

		values, err := getPayloadTemplateValues(exp)
		if err != nil {
			log.Logger.Error("could not get payload template values")
			return "", err
		}

		// get the metrics spec
		var buf bytes.Buffer
		err = template.Execute(&buf, values)
		if err != nil {
			log.Logger.Error("could not execute payload template")
			return "", err
		}

		return buf.String(), nil
	}

	return "", nil
}

// initializeDefaults sets default values for the custom metrics task
func (t *notifyTask) initializeDefaults() {
	// set default HTTP method
	if t.With.Method == "" {
		if t.With.PayloadTemplateURL != "" {
			t.With.Method = "POST"
		} else {
			t.With.Method = "GET"
		}
	}
}

// validate task inputs
func (t *notifyTask) validateInputs() error {
	return nil
}

// run executes this task
func (t *notifyTask) run(exp *Experiment) error {
	// validate inputs
	err := t.validateInputs()
	if err != nil {
		return err
	}

	if exp.driver == nil {
		return errors.New("no driver was provided for collect-metrics-database task")
	}

	// initialize defaults
	t.initializeDefaults()

	var requestBody io.Reader

	log.Logger.Debug("method: ", t.With.Method, " URL: ", t.With.Url)

	if t.With.PayloadTemplateURL != "" {
		payload, err := t.getPayload(exp)
		if err != nil {
			log.Logger.Error("could not get payload")
			return err
		}

		log.Logger.Debug("add payload: ", string(payload))

		requestBody = strings.NewReader(payload)
	}

	// create a new HTTP request
	req, err := http.NewRequest(t.With.Method, t.With.Url, requestBody)
	if err != nil {
		log.Logger.Error("could not create HTTP request for notify task:", err)
		return nil
	}

	// iterate through headers
	for headerName, headerValue := range t.With.Headers {
		req.Header.Add(headerName, headerValue)
		log.Logger.Debug("add header: ", headerName, ", value: ", headerValue)
	}

	// add query params
	q := req.URL.Query()
	for key, value := range t.With.Params {
		q.Add(key, value)
		log.Logger.Debug("add param: ", key, ", value: ", value)
	}
	req.URL.RawQuery = q.Encode()

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Error("could not send HTTP request for notify task:", err)
		return nil
	}
	defer resp.Body.Close()

	// read response responseBody
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error("could not read response body from Slack notification request", err)
		return nil
	}

	log.Logger.Debug("response body: ", string(responseBody))

	return nil
}
