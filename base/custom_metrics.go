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
	"text/template"

	"time"

	"github.com/Masterminds/sprig"
	"github.com/itchyny/gojq"
	log "github.com/iter8-tools/iter8/base/log"
	"sigs.k8s.io/yaml"
)

// ProviderSpec specifies how to get metrics from a provider
type ProviderSpec struct {
	// URL is the database endpoint
	URL string `json:"url" yaml:"url"`

	// Method is the HTTP method that needs to be used
	Method string `json:"method" yaml:"method"`

	// Headers is the set of HTTP headers that need to be sent
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Metrics is the set of metrics that can be obtained
	Metrics []Metric `json:"metrics" yaml:"metrics"`
}

// Metric defines how to construct HTTP requests and parse HTTP responses
// when querying a database for a metric. Metric struct also includes metadata
// such as the name and description of the metric.
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
	Params *[]HTTPParam `json:"params,omitempty" yaml:"params,omitempty"`

	// Body is the HTTP request body that needs to be sent
	Body *string `json:"body,omitempty" yaml:"body,omitempty"`

	// JqExpression is the jq expression that can extract the value from the HTTP
	// response
	JqExpression string `json:"jqExpression" yaml:"jqExpression"`
}

// HTTPParam defines an HTTP parameter
type HTTPParam struct {
	// Name is the name of the HTTP parameter
	Name string `json:"name" yaml:"name"`

	// Value is the value of the HTTP parameter
	Value string `json:"value" yaml:"value"`
}

// customMetricsInputs is the input to the custommetrics task
type customMetricsInputs struct {
	// Templates	Maps provider to its template URL
	Templates map[string]string `json:"templates" yaml:"templates"`

	// Values Values used for substituting placeholders in metric templates.
	Values map[string]interface{} `json:"values" yaml:"values"`

	// VersionValues Per version values that override values
	// For each version, its version values are coalesced with values
	// The length of this slice equals the number of versions
	VersionValues []map[string]interface{} `json:"versionValues" yaml:"versionValues"`
}

const (
	// CustomMetricsTaskName is the name of this task which fetches metrics templates, constructs metric specs, and then fetches metrics for each version from metric provider databases
	CustomMetricsTaskName = "custommetrics"

	// startingTime specifies how far back to go in time for a specific version
	// startingTimeStr is starting time placeholder
	startingTimeStr = "startingTime"

	// how much time has elapsed between startingTime and now
	elapsedTimeSecondsStr = "elapsedTimeSeconds"
)

// customMetricsTask enables collection of custom metrics from databases
type customMetricsTask struct {
	TaskMeta
	With customMetricsInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the custom metrics task
func (t *customMetricsTask) initializeDefaults() {
	// initialize versionValues if absent
	if len(t.With.VersionValues) == 0 {
		t.With.VersionValues = []map[string]interface{}{t.With.Values}
	}
}

// validate task inputs
func (t *customMetricsTask) validateInputs() error {
	return nil
}

// getElapsedTimeSeconds using values and experiment
//
// elapsedTime = currentTime - startingTime
//
// First, check for startingTime in values.
// If unavailable, use startingTime of the experiment.
func getElapsedTimeSeconds(values map[string]interface{}, exp *Experiment) (int64, error) {
	startingTime := exp.Result.StartTime.Time
	if values[startingTimeStr] != nil {
		var err error
		// Calling Parse() method with its parameters
		startingTime, err = time.Parse(time.RFC3339, fmt.Sprintf("%v", values[startingTimeStr]))

		if err != nil {
			return 0, errors.New("cannot parse startingTime")
		}
	}

	// calculate the elapsedTimeSeconds based on the startingTime if it has been provided
	currentTime := time.Now()
	return int64(currentTime.Sub(startingTime).Seconds()), nil
}

// construct request to database and return extracted metric value
//
// bool return value represents whether the pipeline was able to run to
// completion (prevents double error statement)
func queryDatabaseAndGetValue(template ProviderSpec, metric Metric) (interface{}, bool) {
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

// get provider template from URL
func getProviderTemplate(providerURL string) (*template.Template, error) {
	// fetch b from url
	resp, err := http.Get(providerURL)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	// read responseBody
	// get the doubly templated metrics spec
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tpl, err := template.New("provider template").Funcs(sprig.TxtFuncMap()).Parse(string(responseBody))
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return tpl, nil
}

// run executes this task
func (t *customMetricsTask) run(exp *Experiment) error {
	// validate inputs
	var err error

	err = t.validateInputs()
	if err != nil {
		return err
	}

	// initialize defaults
	t.initializeDefaults()

	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionValues))
	if err != nil {
		return err
	}

	// collect metrics from all providers and for all versions
	for providerName, url := range t.With.Templates {
		// finalize metrics spec
		template, err := getProviderTemplate(url)
		if err != nil {
			return err
		}

		for i, versionValues := range t.With.VersionValues {
			// merge values
			vals, err := mustMergeOverwrite(t.With.Values, versionValues)
			if err != nil {
				return err
			}
			values := vals.(map[string]interface{})
			if len(values) == 0 {
				values = make(map[string]interface{})
			}
			// add elapsedTimeSeconds
			elapsedTimeSeconds, err := getElapsedTimeSeconds(values, exp)
			if err != nil {
				return err
			}
			values[elapsedTimeSecondsStr] = elapsedTimeSeconds

			// get the metrics spec
			var buf bytes.Buffer
			err = template.Execute(&buf, values)
			if err != nil {
				return err
			}
			bytes, _ := ioutil.ReadAll(&buf)
			var provider ProviderSpec
			err = yaml.Unmarshal(bytes, &provider)
			if err != nil {
				return err
			}
			log.Logger.Debugf("provider spec %v for version %v\n", providerName, i)
			log.Logger.Debug("--------------------------------")
			log.Logger.Debug(string(bytes))

			// get each metric
			for _, metric := range provider.Metrics {
				log.Logger.Debug("query for metric ", metric.Name)

				// perform database query and extract metric value
				val, ok := queryDatabaseAndGetValue(provider, metric)

				// check if there were any issues querying database and extracting value
				if !ok {
					log.Logger.Error("could not query for metric ", metric.Name, ": ", err)
					continue
				}

				// do not save value if it has no value
				if val == nil {
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
				valueString := fmt.Sprint(val)
				floatValue, err := strconv.ParseFloat(valueString, 64)
				if err != nil {
					log.Logger.Error("could not parse string \""+valueString+"\" to float:", err)
					continue
				}

				err = exp.Result.Insights.updateMetric(providerName+"/"+metric.Name, mm, i, floatValue)

				if err != nil {
					log.Logger.Error("could not add update metric", err)
					continue
				}
			}
		}
	}

	return nil
}
