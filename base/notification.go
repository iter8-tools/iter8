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

	log "github.com/iter8-tools/iter8/base/log"
)

// notificationInputs is the input to the notification task
type notificationInputs struct {
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

	SLOsSatisfied []bool `json:"SLOsSatisfied" yaml:"SLOsSatisfied"`

	Experiment string `json:"experiment" yaml:"experiment"`
}

const (
	// NotificationTaskName is the name of this task which sends notifications
	NotificationTaskName = "notification"

	// NewLine is a newline character
	NewLine string = "\n"

	// Space is a space character
	Space string = " "
)

// notificationTask sends notifications
type notificationTask struct {
	TaskMeta
	With notificationInputs `json:"with" yaml:"with"`
}

type Report struct {
	// Group string `json:"group" yaml:"group"`

	TimeStamp string `json:"timeStamp" yaml:"timeStamp"`

	Completed bool `json:"completed" yaml:"completed"`

	NoTaskFailures bool `json:"noTaskFailures" yaml:"noTaskFailures"`

	NumTasks int `json:"numTasks" yaml:"numTasks"`

	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	NumLoops int `json:"numLoops" yaml:"numLoops"`

	Experiment *Experiment `json:"experiment" yaml:"experiment"`

	// SLOs map[string]SLOReport `json:"SLOs" yaml:"SLOs"`

	// Metrics map[string]MetricReport `json:"metrics" yaml:"metrics"`
}

type NotifyPayloadTemplateValues struct {
	Report Report `json:"report" yaml:"report"`

	JSONStringReport string `json:"textReport" yaml:"textReport"`
}

// type SLOReport struct {
// 	Type MetricType `json:"type" yaml:"type"`

// 	Unit string `json:"unit,omitempty" yaml:"unit,omitempty"`

// 	IsUpper bool `json:"isUpper" yaml:"isUpper"`

// 	Limit float64 `json:"limit" yaml:"limit"`

// 	Satisfied []bool `json:"satisfied" yaml:"satisfied"`
// }

// type MetricReport struct {
// 	Type string `json:"type" yaml:"type"`

// 	Unit string `json:"unit,omitempty" yaml:"unit,omitempty"`

// 	Value string `json:"value" yaml:"value"`
// }

func getPayloadTemplateValues(exp *Experiment) NotifyPayloadTemplateValues {
	// slos := map[string]SLOReport{}

	// for i, upperSlo := range exp.Result.Insights.SLOs.Upper {
	// 	SloMetricMeta, err := exp.Result.Insights.GetMetricsInfo(upperSlo.Metric)

	// 	slos[upperSlo.Metric] = SLOReport{
	// 		Type: SloMetricMeta.Type,

	// 		Unit: *SloMetricMeta.Units,

	// 		IsUpper: true,

	// 		Limit: upperSlo.Limit,

	// 		Satisfied: exp.Result.Insights.SLOsSatisfied.Upper[i],
	// 	}
	// }

	// metrics := map[string]MetricReport{}

	// switch v := exp.driver.(type) {
	// case KubeDriver:
	// 	exp.driver.Group
	// }

	report := Report{
		// Group: exp.driver.Group
		TimeStamp:         time.Now().String(),
		Completed:         exp.Completed(),
		NoTaskFailures:    exp.NoFailure(),
		NumTasks:          len(exp.Spec),
		NumCompletedTasks: exp.Result.NumCompletedTasks,
		NumLoops:          exp.Result.NumLoops,
		Experiment:        exp,

		// SLOs:           slos,
		// Metrics:        metrics,
	}

	jsonReport, err := json.Marshal(report)

	if err != nil {

	}

	return NotifyPayloadTemplateValues{
		Report:           report,
		JSONStringReport: string(jsonReport),
	}
}

// getPayload fetches the payload template from the PayloadTemplateURL and
// executes it with values from getPayloadTemplateValues()
func (t *notificationTask) getPayload(exp *Experiment) string {
	if t.With.PayloadTemplateURL != "" {
		template, err := getProviderTemplate(t.With.PayloadTemplateURL)

		if err != nil {

		}

		values := getPayloadTemplateValues(exp)

		// get the metrics spec
		var buf bytes.Buffer
		err = template.Execute(&buf, values)

		return buf.String()
	}

	return ""
}

// initializeDefaults sets default values for the custom metrics task
func (t *notificationTask) initializeDefaults() {
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
func (t *notificationTask) validateInputs() error {
	return nil
}

// run executes this task
func (t *notificationTask) run(exp *Experiment) error {
	// validate inputs
	var err error

	err = t.validateInputs()
	if err != nil {
		return err
	}

	if exp.driver == nil {
		return errors.New("no driver was provided for collect-metrics-database task")
	}

	// initialize defaults
	t.initializeDefaults()

	// // Reference: https://api.slack.com/reference/messaging/payload
	// slackMessagePayload := map[string]interface{}{
	// 	"text": "Test notification",
	// }

	// var tpl *template.Template

	// var payload []byte

	// if t.With.PayloadTemplateURL != "" {
	// 	template, err := getProviderTemplate(t.With.PayloadTemplateURL)

	// 	if err != nil {

	// 	}

	// 	// get the metrics spec
	// 	var buf bytes.Buffer
	// 	err = template.Execute(&buf, values)
	// }

	// // Marshal Slack payload into JSON
	// slackMessageJson, err := json.Marshal(slackMessagePayload)
	// if err != nil {
	// 	log.Logger.Error("could not JSON marshal Slack notification body:", err)
	// }

	var requestBody io.Reader

	if t.With.PayloadTemplateURL != "" {
		requestBody = strings.NewReader(t.getPayload(exp))
	}

	// create a new HTTP request
	req, err := http.NewRequest(t.With.Method, t.With.Url, requestBody)
	if err != nil {
		log.Logger.Error("could not create HTTP request for notify task:", err)
		return nil
	}

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
