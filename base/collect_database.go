package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
	log "github.com/iter8-tools/iter8/base/log"

	"text/template"
	"time"

	"sigs.k8s.io/yaml"
)

// ToDo: Go Doc is needed in this file

type CollectDatabaseTemplate struct {
	Url      string            `json:"url" yaml:"url"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Provider string            `json:"provider" yaml:"provider"`
	Method   string            `json:"method" yaml:"method"`
	Metrics  []Metric          `json:"metrics" yaml:"metrics"`
}

type Metric struct {
	Name         string    `json:"name" yaml:"name"`
	Description  *string   `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string    `json:"type" yaml:"type"`
	Units        *string   `json:"units,omitempty" yaml:"units,omitempty"`
	Params       *[]Params `json:"params,omitempty" yaml:"params,omitempty"`
	Body         *string   `json:"body,omitempty" yaml:"body,omitempty"`
	JqExpression string    `json:"jqExpression" yaml:"jqExpression"`
}

type Params struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

const startingTimeString = "StartingTime"
const elapsedTimeString = "ElapsedTime"

// ToDo: Iter8 wiki is a great place to document the following (whichs seems IBM Code Engine specific) ... move it there

// collectDatabaseInputs holds all the inputs for this task
//
// Inputs for the template:
//   ibm_codeengine_application_name string
//   ibm_codeengine_gateway_instance string
//   ibm_codeengine_namespace        string
//   ibm_codeengine_project_name     string
//   ibm_codeengine_revision_name    string
//   ibm_codeengine_status           string
//   ibm_ctype                       string
//   ibm_location                    string
//   ibm_scope                       string
//   ibm_service_instance            string
//   ibm_service_name                string
//
// Inputs for the metrics (output of template):
//   ibm_codeengine_revision_name string
//   StartingTime                 int64 (UNIX time stamp)
//
// Note: ElapsedTime is produced by Iter8
type collectDatabaseInputs struct {
	VersionInfo []map[string]interface{} `json:"versionInfo" yaml:"versionInfo"`
}

const (
	// collectDatabaseTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectDatabaseTaskName = "collect-metrics-database"
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
func getElapsedTime(versionInfo map[string]interface{}, exp *Experiment) (int64, error) {
	// ElapsedTime should not be provided by the user
	if versionInfo[elapsedTimeString] != nil {
		return 0, errors.New("ElapsedTime should not be provided by the user in VersionInfo: " + fmt.Sprintf("%v", versionInfo))
	}

	// set StartingTime based on VersionInfo or start of the experiment
	var startingTime int64
	if versionInfo[startingTimeString] != nil {
		rawStartingTime := versionInfo[startingTimeString]
		switch rawStartingTime := rawStartingTime.(type) {
		case int64:
			startingTime = rawStartingTime
		case int:
		case float64: // parsing the metrics file gives float64
			startingTime = int64(rawStartingTime)
		default:
			return 0, errors.New("Cannot integer parse StartingTime from VersionInfo: " + fmt.Sprintf("%v", versionInfo))
		}
	} else {
		startingTime = exp.Result.StartTime.Unix()
	}

	// calculate the ElapsedTime based on the StartingTime if it has been provided
	currentTime := time.Now().Unix()
	return currentTime - startingTime, nil
}

// construct request to database and return extracted metric value
//
// bool return value represents whether the pipeline was able to run to
// completion (prevents double error statement)
func queryDatabaseAndGetValue(template CollectDatabaseTemplate, metric Metric) (interface{}, bool) {
	var requestBody io.Reader
	if metric.Body != nil {
		requestBody = strings.NewReader(*metric.Body)
	} else {
		requestBody = nil
	}

	// create a new HTTP request
	req, err := http.NewRequest(template.Method, template.Url, requestBody)
	if err != nil {
		log.Logger.Error("could not create new request for metric ", metric.Name, ": ", err)
		return nil, false
	}

	// iterate through headers
	for headerName, headerValue := range template.Headers {
		req.Header.Add(headerName, headerValue)
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")

	// add query params
	q := req.URL.Query()
	params := metric.Params
	for _, param := range *params {
		q.Add(param.Name, param.Value)
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

	// initialize defaults
	t.initializeDefaults()

	// get current directory path
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	// collect all files paths in current directory that ends with metrics.yaml
	metricFilePaths := []string{}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, "metrics.yaml") {
			metricFilePaths = append(metricFilePaths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// inputs for this task determine the number of versions participating in the
	// experiment. Initiate insights with num versions.
	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionInfo))
	if err != nil {
		return err
	}

	// collect metrics for all metric files and versionInfos
	for _, metricFilePath := range metricFilePaths {
		for i, versionInfo := range t.With.VersionInfo {
			// add ElapsedTime
			elapsedTime, err := getElapsedTime(versionInfo, exp)
			if err != nil {
				return err
			}
			versionInfo[elapsedTimeString] = elapsedTime

			// finalize metrics template
			template, err := template.ParseFiles(metricFilePath)
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			err = template.Execute(&buf, versionInfo)
			if err != nil {
				return err
			}
			var metrics CollectDatabaseTemplate
			err = yaml.Unmarshal(buf.Bytes(), &metrics)
			if err != nil {
				return err
			}

			for _, metric := range metrics.Metrics {
				// perform database query and extract metric value
				value, ok := queryDatabaseAndGetValue(metrics, metric)

				// check if there were any issues querying database and extracting value
				if !ok {
					continue
				}

				// do not save value if it has no value
				if value == nil {
					log.Logger.Error("could not extract non-nil value for metric ", metric.Name, ": ", err)
					continue
				}

				// determine metric name
				pathTokens := strings.Split(metricFilePath, "/")
				fileNameWithExtension := pathTokens[len(pathTokens)-1]
				fileNameTokens := strings.Split(fileNameWithExtension, ".metrics.yaml")
				fileName := fileNameTokens[0]
				metricName := fileName + "/" + metric.Name

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

				err = exp.Result.Insights.updateMetric(metricName, mm, i, floatValue)

				if err != nil {
					log.Logger.Error("could not add update metric", err)
					continue
				}
			}
		}
	}

	return nil
}
