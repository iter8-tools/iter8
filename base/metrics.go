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
	// PerformanceResultPath is the path to the PUT performanceResult/ endpoint
	PerformanceResultPath = "/performanceResult"

	// HTTPDashboardPath is the path to the GET httpDashboard/ endpoint
	HTTPDashboardPath = "/httpDashboard"
	// GRPCDashboardPath is the path to the GET grpcDashboard/ endpoint
	GRPCDashboardPath = "/grpcDashboard"
)

func putPerformanceResultToMetricsService(metricsServerURL, namespace, experiment string, data interface{}) error {
	// handle URL and URL parameters
	u, err := url.ParseRequestURI(metricsServerURL + PerformanceResultPath)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("namespace", namespace)
	params.Add("experiment", experiment)
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	log.Logger.Trace(fmt.Sprintf("performance result URL: %s", urlStr))

	// handle payload
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Logger.Error("cannot JSON marshal data for metrics server request: ", err)
		return err
	}

	// create request
	req, err := http.NewRequest(http.MethodPut, urlStr, bytes.NewBuffer(dataBytes))
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
