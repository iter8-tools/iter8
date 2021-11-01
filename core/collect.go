package core

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
	log "github.com/iter8-tools/iter8/core/log"
)

const (
	// TaskName is the name of the task this file implements
	CollectTaskName = "collect-fortio-metrics"

	// DefaultQPS is the default value of QPS (queries per sec) in the collect task
	DefaultQPS = float32(8)

	// DefaultNumRequests is the default value of the number of requests used by the collect task
	DefaultNumRequests = int64(100)

	// DefaultConnections is the default value of the number of connections
	DefaultConnections = uint32(4)

	// RequestCountMetricName is the name of the request-count metric
	RequestCountMetricName = "request-count"

	// ErrorCountMetricName is the name of the error-count metric
	ErrorCountMetricName = "error-count"

	// ErrorRateMetricName is the name of the error-rate metric
	ErrorRateMetricName = "error-rate"

	// MeanLatencyMetricName is the name of the mean-latency metric
	MeanLatencyMetricName = "mean-latency"

	// StdDevMetricName is the name of the latency standard deviation metric
	StdDevMetricName = "stddev-latency"

	// MinLatencyMetricName is the name of the min-latency metric
	MinLatencyMetricName = "min-latency"

	// MaxLatencyMetricName is the name of the max-latency metric
	MaxLatencyMetricName = "max-latency"
)

var (
	// DefaultErrorRanges is the default value of the error ranges
	DefaultErrorRanges = []ErrorRange{{Lower: IntPointer(500)}}

	// DefaultPercentiles is the default value for latency percentiles
	DefaultPercentiles = [...]float64{50.0, 75.0, 90.0, 95.0, 99.0, 99.9}
)

// Version contains header and url information needed to send requests to each version.
type Version struct {
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers" yaml:"headers"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// HTTP status code within this range is considered an error
type ErrorRange struct {
	Lower *int `json:"lower" yaml:"lower"`
	Upper *int `json:"upper" yaml:"upper"`
}

// CollectInputs contain the inputs to the metrics collection task to be executed.
type CollectInputs struct {
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
	ErrorRanges []ErrorRange `json:"errorRanges" yaml:"errorRanges"`
	// Percentiles are the set of latency percentiles to be collected
	Percentiles []float64 `json:"percentiles" yaml:"percentiles"`
	// information about versions
	VersionInfo []*Version `json:"versionInfo" yaml:"versionInfo"`
}

// ErrorCode checks if a given code is an error code
func (t *CollectTask) ErrorCode(code int) bool {
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

// CollectTask enables collection of Iter8's built-in metrics.
type CollectTask struct {
	TaskMeta
	With CollectInputs `json:"with" yaml:"with"`
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
		ct := &CollectTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		if ct.With.VersionInfo == nil {
			return nil, errors.New("collect task with nil versionInfo")
		}
		bt = ct
	}
	return bt, err
}

// InitializeDefaults sets default values for the collect task
func (t *CollectTask) InitializeDefaults() {
	if t.With.NumRequests == nil && t.With.Duration == nil {
		t.With.NumRequests = Int64Pointer(DefaultNumRequests)
	}
	if t.With.QPS == nil {
		t.With.QPS = Float32Pointer(DefaultQPS)
	}
	if t.With.ErrorRanges == nil {
		t.With.ErrorRanges = DefaultErrorRanges
	}
	if t.With.Percentiles == nil {
		for _, p := range DefaultPercentiles {
			t.With.Percentiles = append(t.With.Percentiles, p)
		}
	}
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func (t *CollectTask) getFortioOptions(j int) (*fhttp.HTTPRunnerOptions, error) {
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
func (t *CollectTask) resultForVersion(j int) (*fhttp.HTTPRunnerResults, error) {
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
func (t *CollectTask) Run(exp *Experiment) error {
	var err error
	t.InitializeDefaults()

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
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+RequestCountMetricName, i, float64(fm[i].DurationHistogram.Count))
			if err != nil {
				return err
			}

			// error count and rate
			val := float64(0)
			for code, count := range fm[i].RetCodes {
				if t.ErrorCode(code) {
					val += float64(count)
				}
			}
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+ErrorCountMetricName, i, val)
			if err != nil {
				return err
			}

			// error-rate
			rc := float64(fm[i].DurationHistogram.Count)
			if rc != 0 {
				err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+ErrorRateMetricName, i, val/rc)
				if err != nil {
					return err
				}
			}

			// mean-latency
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+MeanLatencyMetricName, i, fm[i].DurationHistogram.Avg)
			if err != nil {
				return err
			}

			// stddev-latency
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+StdDevMetricName, i, fm[i].DurationHistogram.StdDev)
			if err != nil {
				return err
			}

			// min-latency
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+MinLatencyMetricName, i, fm[i].DurationHistogram.Min)
			if err != nil {
				return err
			}

			// max-latency
			err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+MaxLatencyMetricName, i, fm[i].DurationHistogram.Max)
			if err != nil {
				return err
			}

			for _, p := range fm[i].DurationHistogram.Percentiles {
				err = exp.UpdateMetricForVersion(Iter8FortioPrefix+"/"+fmt.Sprintf("%0.2f", p.Percentile), i, p.Value)
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
