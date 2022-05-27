package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"time"

	"github.com/itchyny/gojq"
	log "github.com/iter8-tools/iter8/base/log"

	"sigs.k8s.io/yaml"
)

// MetricsSpec specifies the set of metrics that can be obtained from a provider
type MetricsSpec struct {
	// Provider is the label/name of the data source
	Provider string `json:"provider" yaml:"provider"`

	// URL is the database endpoint
	URL string `json:"url" yaml:"url"`

	// Method is the HTTP method that needs to be used
	Method string `json:"method" yaml:"method"`

	// Headers is the set of HTTP headers that need to be sent
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Metrics is the set of metrics that can be obtained
	Metrics []Metric `json:"metrics" yaml:"metrics"`
}

// Metric defines how to obtain a metric
type Metric struct {
	// Name is the name of the metric
	Name string `json:"name" yaml:"name"`

	// Description is the description of the metric
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Type is the type of the metric, either gauge or counter
	Type string `json:"type" yaml:"type"`

	// Units is the unit of the metric, which can be omitted for unitless metrics
	Units *string `json:"units,omitempty" yaml:"units,omitempty"`

	// Params is the set of HTTP parameters that need to be sent
	Params *[]Params `json:"params,omitempty" yaml:"params,omitempty"`

	// Body is the HTTP request body that needs to be sent
	Body *string `json:"body,omitempty" yaml:"body,omitempty"`

	// JqExpression is the jq expression that can extract the value from the HTTP
	// response
	JqExpression string `json:"jqExpression" yaml:"jqExpression"`
}

// Params defines an HTTP parameter
type Params struct {
	// Name is the name of the HTTP parameter
	Name string `json:"name" yaml:"name"`

	// Value is the value of the HTTP parameter
	Value string `json:"value" yaml:"value"`
}

// collectDatabaseInputs is the input to the collect-metrics-database task
type collectDatabaseInputs struct {
	// Providers is the set of labels/names of the data sources
	Providers []string `json:"providers" yaml:"providers"`

	// VersionInfo
	VersionInfo []map[string]interface{} `json:"versionInfo" yaml:"versionInfo"`
}

const (
	// CollectDatabaseTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectDatabaseTaskName = "collect-metrics-database"

	// experimentMetricsPathSuffix is the name of the metrics spec file
	experimentMetricsPathSuffix = ".metrics.yaml"

	startingTimeStr = "startingTime"

	elapsedTimeSecondsStr = "elapsedTimeSeconds"

	// timeLayout is an example time layout for startingTime
	timeLayout = "Jan 2, 2006 at 3:04pm (MST)"
)

// collectDatabaseTask enables load testing of gRPC services.
type collectDatabaseTask struct {
	TaskMeta
	With collectDatabaseInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectDatabaseTask) initializeDefaults() {
}

// validate task inputs
func (t *collectDatabaseTask) validateInputs() error {
	return nil
}

// for a given version info and Experiment, calculate the elapsed time that
// should be used for queries
//
// elapsed time is based on the StartingTime in the version info or the
// starting time in the Experiment
func getElapsedTimeSeconds(versionInfo map[string]interface{}, exp *Experiment) (int64, error) {
	// elapsedTimeSeconds should not be provided by the user
	if versionInfo[elapsedTimeSecondsStr] != nil {
		return 0, errors.New("elapsedTimeSeconds should not be provided by the user in VersionInfo: " + fmt.Sprintf("%v", versionInfo))
	}

	startingTime := exp.Result.StartTime.Unix()
	if versionInfo[startingTimeStr] != nil {
		// Calling Parse() method with its parameters
		temp, err := time.Parse(timeLayout, fmt.Sprintf("%v", versionInfo[startingTimeStr]))

		if err != nil {
			return 0, errors.New("cannot parse startingTime")
		} else {
			startingTime = temp.Unix()
		}
	}

	// calculate the elapsedTimeSeconds based on the startingTime if it has been provided
	currentTime := time.Now().Unix()
	return currentTime - startingTime, nil
}

// construct request to database and return extracted metric value
//
// bool return value represents whether the pipeline was able to run to
// completion (prevents double error statement)
func queryDatabaseAndGetValue(template MetricsSpec, metric Metric) (interface{}, bool) {
	var requestBody io.Reader
	if metric.Body != nil {
		requestBody = strings.NewReader(*metric.Body)
	} else {
		requestBody = nil
	}

	// create a new HTTP request
	req, err := http.NewRequest(template.Method, template.URL, requestBody)
	if err != nil {
		log.Logger.Error("could not create new request for metric ", metric.Name, ": ", err)
		return nil, false
	}

	// iterate through headers
	for headerName, headerValue := range template.Headers {
		req.Header.Add(headerName, headerValue)
		log.Logger.Debug("add header: ", headerName, ", value: ", headerValue)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	// add query params
	q := req.URL.Query()
	params := metric.Params
	for _, param := range *params {
		q.Add(param.Name, param.Value)
		log.Logger.Debug("add param: ", param.Name, ", value: ", param.Value)
	}
	req.URL.RawQuery = q.Encode()

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Error("could not request metric ", metric.Name, ": ", err)
		return nil, false
	}
	defer resp.Body.Close()

	// read response responseBody
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error("could not read response body for metric ", metric.Name, ": ", err)
		return nil, false
	}

	log.Logger.Debug("response body: ", string(responseBody))

	// JSON parse response body
	var jsonBody interface{}
	err = json.Unmarshal([]byte(responseBody), &jsonBody)
	if err != nil {
		log.Logger.Error("could not JSON parse response body for metric ", metric.Name, ": ", err)
		return nil, false
	}

	// perform jq expression
	query, err := gojq.Parse(metric.JqExpression)
	if err != nil {
		log.Logger.Error("could not parse jq expression \""+metric.JqExpression+"\" for metric ", metric.Name, ": ", err)
		return nil, false
	}
	iter := query.Run(jsonBody)

	value, ok := iter.Next()
	if !ok {
		log.Logger.Error("could not extract value with jq expression for metric ", metric.Name, ": ", err)
		return nil, false
	}

	return value, true
}

// run executes this task
func (t *collectDatabaseTask) run(exp *Experiment) error {
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

	// inputs for this task determine the number of versions participating in the
	// experiment. Initiate insights with num versions.
	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionInfo))
	if err != nil {
		return err
	}

	// collect metrics for all metric files and versionInfos
	for _, provider := range t.With.Providers {
		for i, versionInfo := range t.With.VersionInfo {
			// add elapsedTimeSeconds
			elapsedTimeSeconds, err := getElapsedTimeSeconds(versionInfo, exp)
			if err != nil {
				return err
			}
			versionInfo[elapsedTimeSecondsStr] = elapsedTimeSeconds

			// finalize metrics spec
			template, err := exp.driver.ReadMetricsSpec(provider)
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			err = template.Execute(&buf, versionInfo)
			if err != nil {
				return err
			}
			var metrics MetricsSpec
			err = yaml.Unmarshal(buf.Bytes(), &metrics)
			if err != nil {
				return err
			}

			for _, metric := range metrics.Metrics {
				log.Logger.Debug("query for metric ", metric.Name)

				// perform database query and extract metric value
				value, ok := queryDatabaseAndGetValue(metrics, metric)

				// check if there were any issues querying database and extracting value
				if !ok {
					log.Logger.Error("could not query for metric ", metric.Name, ": ", err)
					continue
				}

				// do not save value if it has no value
				if value == nil {
					log.Logger.Error("could not extract non-nil value for metric ", metric.Name, ": ", err)
					continue
				}

				// determine metric type
				var metricType MetricType
				if metric.Type == "gauge" {
					metricType = GaugeMetricType
				} else if metric.Type == "counter" {
					metricType = CounterMetricType
				}

				// finalize metric data
				mm := MetricMeta{
					Description: *metric.Description,
					Type:        metricType,
					Units:       metric.Units,
				}

				// convert value to float
				valueString := fmt.Sprint(value)
				floatValue, err := strconv.ParseFloat(valueString, 64)
				if err != nil {
					log.Logger.Error("could not parse string \""+valueString+"\" to float:", err)
					continue
				}

				err = exp.Result.Insights.updateMetric(provider+"/"+metric.Name, mm, i, floatValue)

				if err != nil {
					log.Logger.Error("could not add update metric", err)
					continue
				}
			}
		}
	}

	return nil
}
