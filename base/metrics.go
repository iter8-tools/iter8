package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/iter8-tools/iter8/base/log"
)

const (
	// MetricsServerURL is the URL of the metrics server
	MetricsServerURL = "METRICS_SERVER_URL"

	// ExperimentResultPath is the path to the PUT /experimentResult endpoint
	ExperimentResultPath = "/experimentResult"

	// AbnDashboard is the path to the GET /abnDashboard endpoint
	AbnDashboard = "/abnDashboard"
	// HTTPDashboardPath is the path to the GET /httpDashboard endpoint
	HTTPDashboardPath = "/httpDashboard"
	// GRPCDashboardPath is the path to the GET /grpcDashboard endpoint
	GRPCDashboardPath = "/grpcDashboard"
)

// callMetricsService is a general function that can be used to send data to the metrics service
func callMetricsService(method, metricsServerURL, path string, queryParams map[string]string, payload interface{}) error {
	// handle URL and URL parameters
	u, err := url.ParseRequestURI(metricsServerURL + path)
	if err != nil {
		return err
	}

	params := url.Values{}
	for paramKey, paramValue := range queryParams {
		params.Add(paramKey, paramValue)
	}
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	log.Logger.Trace(fmt.Sprintf("call metrics service URL: %s", urlStr))

	// handle payload
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		log.Logger.Error("cannot JSON marshal data for metrics server request: ", err)
		return err
	}

	// create request
	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(dataBytes))
	if err != nil {
		log.Logger.Error("cannot create new HTTP request metrics server: ", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	log.Logger.Trace("sending request")

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Error("could not send request to metrics server: ", err)
		return err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Logger.Error("could not close response body: ", err)
		}
	}()

	log.Logger.Trace("sent request")

	return nil
}

// PutExperimentResultToMetricsService sends the test result to the metrics service
func PutExperimentResultToMetricsService(metricsServerURL, namespace, test string, testResult *ExperimentResult) error {
	return callMetricsService(http.MethodPut, metricsServerURL, ExperimentResultPath, map[string]string{
		"namespace": namespace,
		"test":      test,
	}, testResult)
}
