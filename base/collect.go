package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"fortio.org/fortio/fhttp"
	fortioLog "fortio.org/fortio/log"
	"fortio.org/fortio/periodic"
	log "github.com/iter8-tools/iter8/base/log"
)

// version contains header and url information needed to send requests to each version.
type version struct {
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers" yaml:"headers"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// HTTP status code within this range is considered an error
type errorRange struct {
	Lower *int `json:"lower" yaml:"lower"`
	Upper *int `json:"upper" yaml:"upper"`
}

// collectInputs contain the inputs to the metrics collection task to be executed.
type collectInputs struct {
	// how many requests will be sent for each version; optional; default 100
	NumRequests *int64 `json:"numRequests" yaml:"numRequests"`
	// how long to run the metrics collector; optional;
	// if both time and numRequests are specified, numRequests takes precedence
	Duration *string `json:"time" yaml:"time"`
	// how many queries per second will be sent; optional; default 8
	QPS *float32 `json:"qps" yaml:"qps"`
	// string to be sent during queries as payload; optional
	PayloadStr *string `json:"payloadStr" yaml:"payloadStr"`
	// URL whose content will be sent as payload during queries; optional
	// if both payloadURL and payloadStr are specified, the URL takes precedence
	PayloadURL *string `json:"payloadURL" yaml:"payloadURL"`
	// valid HTTP content type string; specifying this switches the request from GET to POST
	ContentType *string `json:"contentType" yaml:"contentType"`
	// ErrorRanges of HTTP status codes that are considered as errors
	ErrorRanges []errorRange `json:"errorRanges" yaml:"errorRanges"`
	// Percentiles are the set of latency percentiles to be collected
	Percentiles []float64 `json:"percentiles" yaml:"percentiles"`
	// information about versions
	VersionInfo []*version `json:"versionInfo" yaml:"versionInfo"`
}

const (
	CollectTaskName        = "collect-fortio-metrics"
	defaultQPS             = float32(8)
	defaultNumRequests     = int64(100)
	defaultConnections     = uint32(4)
	requestCountMetricName = "request-count"
	errorCountMetricName   = "error-count"
	errorRateMetricName    = "error-rate"
	meanLatencyMetricName  = "mean-latency"
	stdDevMetricName       = "stddev-latency"
	minLatencyMetricName   = "min-latency"
	maxLatencyMetricName   = "max-latency"
)

var (
	defaultErrorRanges = []errorRange{{Lower: intPointer(500)}}
	defaultPercentiles = [...]float64{50.0, 75.0, 90.0, 95.0, 99.0, 99.9}
)

// errorCode checks if a given code is an error code
func (t *collectTask) errorCode(code int) bool {
	for _, lims := range t.With.ErrorRanges {
		// if no lower limit (check upper)
		if lims.Lower == nil && code <= *lims.Upper {
			return true
		}
		// if no upper limit (check lower)
		if lims.Upper == nil && code >= *lims.Lower {
			return true
		}
		// if both limits are present (check both)
		if lims.Upper != nil && lims.Lower != nil && code <= *lims.Upper && code >= *lims.Lower {
			return true
		}
	}
	return false
}

// collectTask enables collection of Iter8's built-in metrics.
type collectTask struct {
	taskMeta
	With collectInputs `json:"with" yaml:"with"`
}

// MakeCollect constructs a CollectTask out of a collect task spec
func MakeCollect(t *TaskSpec) (Task, error) {
	if *t.Task != CollectTaskName {
		return nil, errors.New("task need to be " + CollectTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to CollectTask
	if err == nil {
		ct := &collectTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		if ct.With.VersionInfo == nil {
			return nil, errors.New("collect task with nil versionInfo")
		}
		bt = ct
	}
	return bt, err
}

// initializeDefaults sets default values for the collect task
func (t *collectTask) initializeDefaults() {
	if t.With.NumRequests == nil && t.With.Duration == nil {
		t.With.NumRequests = int64Pointer(defaultNumRequests)
	}
	if t.With.QPS == nil {
		t.With.QPS = float32Pointer(defaultQPS)
	}
	if t.With.ErrorRanges == nil {
		t.With.ErrorRanges = defaultErrorRanges
	}
	if t.With.Percentiles == nil {
		for _, p := range defaultPercentiles {
			t.With.Percentiles = append(t.With.Percentiles, p)
		}
	}
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func (t *collectTask) getFortioOptions(j int) (*fhttp.HTTPRunnerOptions, error) {
	fortioLog.SetOutput(io.Discard)
	// basic runner
	fo := &fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			RunType:     "Iter8 load test",
			QPS:         float64(*t.With.QPS),
			Percentiles: t.With.Percentiles,
			Out:         io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: t.With.VersionInfo[j].URL,
		},
	}

	//num requests
	if t.With.NumRequests != nil {
		fo.RunnerOptions.Exactly = *t.With.NumRequests
	}

	// add duration
	var duration time.Duration
	var err error
	if t.With.Duration != nil {
		duration, err = time.ParseDuration(*t.With.Duration)
		if err == nil {
			fo.RunnerOptions.Duration = duration
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("unable to parse duration")
			return nil, err
		}
	}

	// content type & payload
	if t.With.ContentType != nil {
		fo.ContentType = *t.With.ContentType
	}
	if t.With.PayloadStr != nil {
		fo.Payload = []byte(*t.With.PayloadStr)
	}
	if t.With.PayloadURL != nil {
		b, err := GetPayloadBytes(*t.With.PayloadURL)
		if err != nil {
			return nil, err
		} else {
			fo.Payload = b
		}
	}

	// headers
	for key, value := range t.With.VersionInfo[j].Headers {
		fo.AddAndValidateExtraHeader(key + ":" + value)
	}

	return fo, nil
}

// resultForVersion collects Fortio result for a given version
func (t *collectTask) resultForVersion(j int) (*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio shell command with proper args
	// collect Fortio output as a file
	// and extract the result from the file, and return the result

	fo, err := t.getFortioOptions(j)
	if err != nil {
		return nil, err
	}
	ifr, err := fhttp.RunHTTPTest(fo)
	return ifr, err
}

// Run executes the metrics/collect task
func (t *collectTask) Run(exp *Experiment) error {
	var err error
	t.initializeDefaults()

	fm := make([]*fhttp.HTTPRunnerResults, len(exp.Spec.Versions))

	// run fortio queries for each version sequentially
	for j := range t.With.VersionInfo {
		var data *fhttp.HTTPRunnerResults
		var err error
		if t.With.VersionInfo[j] == nil {
			data = nil
		} else {
			data, err = t.resultForVersion(j)
			if err == nil {
				fm[j] = data
			} else {
				return err
			}
		}
	}

	// set metrics for each version for which fortio metrics are available
	for i := range exp.Spec.Versions {
		if fm[i] != nil {

			// request count
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+requestCountMetricName, i, float64(fm[i].DurationHistogram.Count))
			if err != nil {
				return err
			}

			// error count and rate
			val := float64(0)
			for code, count := range fm[i].RetCodes {
				if t.errorCode(code) {
					val += float64(count)
				}
			}
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+errorCountMetricName, i, val)
			if err != nil {
				return err
			}

			// error-rate
			rc := float64(fm[i].DurationHistogram.Count)
			if rc != 0 {
				err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+errorRateMetricName, i, val/rc)
				if err != nil {
					return err
				}
			}

			// mean-latency
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+meanLatencyMetricName, i, fm[i].DurationHistogram.Avg)
			if err != nil {
				return err
			}

			// stddev-latency
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+stdDevMetricName, i, fm[i].DurationHistogram.StdDev)
			if err != nil {
				return err
			}

			// min-latency
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+minLatencyMetricName, i, fm[i].DurationHistogram.Min)
			if err != nil {
				return err
			}

			// max-latency
			err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+maxLatencyMetricName, i, fm[i].DurationHistogram.Max)
			if err != nil {
				return err
			}

			for _, p := range fm[i].DurationHistogram.Percentiles {
				err = exp.updateMetricForVersion(iter8FortioPrefix+"/"+fmt.Sprintf("p%0.1f", p.Percentile), i, p.Value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// GetPayloadBytes downloads payload from URL and returns a byte slice
func GetPayloadBytes(url string) ([]byte, error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		return nil, errors.New("error while fetching payload")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, err
}
