package base

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// collectCEInputs holds all the inputs for this task
type collectCEInputs struct {
	// IBMInstanceId is the GUID of instance of the IBM Cloud Monitoring instance
	// # https://cloud.ibm.com/docs/monitoring?topic=monitoring-mon-curl
	IBMInstanceId *string `json:"IBMInstanceId" yaml:"IBMInstanceId"`
	// CESysdigUrl is the URL to the CE application
	CESysdigUrl *string `json:"CESysdigUrl" yaml:"CESysdigUrl"`
	// CESysdigToken is the IAM token, used to authenticate the IBM Cloud Monitoring service
	// Related: https://cloud.ibm.com/docs/monitoring?topic=monitoring-mon-curl#mon-curl-headers-iam
	CESysdigToken *string `json:"CESysdigToken" yaml:"CESysdigToken"`
	// URL is the URL to the monitoring isntance
	URL *string `json:"URL" yaml:"URL"`
}

const (
	// collectCETaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectCETaskName = "collect-metrics-ce"
)

// collectCETask enables load testing of gRPC services.
type collectCETask struct {
	TaskMeta
	With collectCEInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectCETask) initializeDefaults() {
}

// validate task inputs
func (t *collectCETask) validateInputs() error {
	return nil
}

// Run executes this task
func (t *collectCETask) Run(exp *Experiment) error {
	// 1. validate inputs
	var err error

	err = t.validateInputs()
	if err != nil {
		return err
	}

	// 2. initialize defaults
	t.initializeDefaults()

	var postBody = `{"metrics": [{"id": "ibm_codeengine_application_revision_count"}], "filter": "ibm_location = \"ca-tor\"", "sampling": 86400, "last": 86400}`

	// Create a new request using http
	req, err := http.NewRequest("POST", *t.With.URL, strings.NewReader(postBody))

	// Add authorization header
	var bearer = "Bearer " + *t.With.CESysdigToken
	req.Header.Add("Authorization", bearer)

	// Add GUID header
	req.Header.Add("IBMInstanceID", *t.With.IBMInstanceId)

	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}

	fmt.Println(string([]byte(body)))

	log.Println(string([]byte(body)))

	return errors.New("not implemented")
}
