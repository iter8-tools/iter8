package base

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/iter8-tools/iter8/base/log"
)

const slackHook = "https://hooks.slack.com/services/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"

// notificationInputs is the input to the notification task
type notificationInputs struct {
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

// initializeDefaults sets default values for the custom metrics task
func (t *notificationTask) initializeDefaults() {
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

	// Reference: https://api.slack.com/reference/messaging/payload
	slackMessagePayload := map[string]interface{}{
		"text": "Test notification",
	}

	// Marshal Slack payload into JSON
	slackMessageJson, err := json.Marshal(slackMessagePayload)
	if err != nil {
		log.Logger.Error("could not JSON marshal Slack notification body:", err)
	}

	// create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, slackHook, strings.NewReader(string(slackMessageJson)))
	if err != nil {
		log.Logger.Error("could not create new request for Slack notification:", err)
		return nil
	}

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Error("could not send Slack notification:", err)
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
