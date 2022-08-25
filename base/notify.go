package base

import (
	"bytes"
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

// getReport gets the values for the payload tempalte
func getReport(exp *Experiment) map[string]Report {
	return map[string]Report{
		"Report": {
			TimeStamp:         time.Now().String(),
			Completed:         exp.Completed(),
			NoTaskFailures:    exp.NoFailure(),
			NumTasks:          len(exp.Spec),
			NumCompletedTasks: exp.Result.NumCompletedTasks,
			NumLoops:          exp.Result.NumLoops,
			Experiment:        exp,
		},
	}
}

// getPayload fetches the payload template from the PayloadTemplateURL and
// executes it with values from getReport()
func (t *notifyTask) getPayload(exp *Experiment) (string, error) {
	if t.With.PayloadTemplateURL != "" {
		template, err := getTextTemplateFromURL(t.With.PayloadTemplateURL)
		if err != nil {
			return "", err
		}

		values := getReport(exp)

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
	if t.With.Url == "" {
		return errors.New("no URL was provided for notify task")
	}

	return nil
}

// run executes this task
func (t *notifyTask) run(exp *Experiment) error {
	// validate inputs
	err := t.validateInputs()
	if err != nil {
		return err
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

		if t.With.SoftFailure {
			return nil
		} else {
			return err
		}
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

		if t.With.SoftFailure {
			return nil
		} else {
			return err
		}
	}
	defer resp.Body.Close()

	if !t.With.SoftFailure && (resp.StatusCode < 200 || resp.StatusCode > 299) {
		return errors.New("did not receive successful status code for notify task")
	}

	// read response responseBody
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error("could not read response body from notification request", err)

		return nil
	}

	log.Logger.Debug("response body: ", string(responseBody))

	return nil
}
