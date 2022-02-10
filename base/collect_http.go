package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"fortio.org/fortio/fhttp"
	fortioLog "fortio.org/fortio/log"
	"fortio.org/fortio/stats"
	log "github.com/iter8-tools/iter8/base/log"
	"github.com/jinzhu/copier"
)

// collectHTTPInputs contain the inputs to the metrics collection task to be executed.
type collectHTTPInputs struct {
	fhttp.HTTPRunnerOptions
	// PayloadJSON is the JSON data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to versions using this data as the JSON payload.
	PayloadJSON interface{} `json:"payloadJSON" yaml:"payloadJSON"`
	// PayloadURL is the URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions using data downloaded from this URL as the payload. If both `payloadStr` and `payloadURL` is specified, the former is ignored.
	PayloadURL *string `json:"payloadURL" yaml:"payloadURL"`
	// ErrorsAbove is the value at or above which HTTP response codes are considered as errors.
	ErrorsAbove *int `json:"errorsAbove" yaml:"errorsAbove"`
	// VersionInfo is a non-empty list of version values.
	VersionInfo []*fhttp.HTTPRunnerOptions `json:"versionInfo" yaml:"versionInfo"`
}

const (
	// CollectHTTPTaskName is the name of this task which performs load generation and metrics collection.
	CollectHTTPTaskName                = "gen-load-and-collect-metrics-http"
	defaultErrorsAbove                 = 400
	builtInHTTPRequestCountId          = "http-request-count"
	builtInHTTPErrorCountId            = "http-error-count"
	builtInHTTPErrorRateId             = "http-error-rate"
	builtInHTTPLatencyMeanId           = "http-latency-mean"
	builtInHTTPLatencyStdDevId         = "http-latency-stddev"
	builtInHTTPLatencyMinId            = "http-latency-min"
	builtInHTTPLatencyMaxId            = "http-latency-max"
	builtInHTTPLatencyHistId           = "http-latency"
	builtInHTTPLatencyPercentilePrefix = "http-latency-p"
	contentTypeJSON                    = "application/json"
)

var (
	defaultPercentiles = [...]float64{50.0, 75.0, 90.0, 95.0, 99.0, 99.9}
)

// errorCode checks if a given code is an error code
func (t *collectHTTPTask) errorCode(code int) bool {
	return code >= *t.With.ErrorsAbove
}

// collectHTTPTask enables load testing of HTTP services.
type collectHTTPTask struct {
	TaskMeta
	With collectHTTPInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectHTTPTask) initializeDefaults() {
	if t.With.ErrorsAbove == nil {
		t.With.ErrorsAbove = intPointer(defaultErrorsAbove)
	}
	// default percentiles are always collected
	// if other percentiles are specified, they are collected as well
	for _, p := range defaultPercentiles {
		t.With.Percentiles = append(t.With.Percentiles, p)
	}
	tmp := Uniq(t.With.Percentiles)
	t.With.Percentiles = []float64{}
	for _, val := range tmp {
		t.With.Percentiles = append(t.With.Percentiles, val.(float64))
	}
}

//validateInputs for this task
func (t *collectHTTPTask) validateInputs() error {
	return nil
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func (t *collectHTTPTask) getFortioOptions(j int) (*fhttp.HTTPRunnerOptions, error) {
	fortioLog.SetOutput(io.Discard)
	// base options
	fo := &fhttp.HTTPRunnerOptions{}
	copier.Copy(fo, &t.With.HTTPRunnerOptions)
	fo.URL = t.With.VersionInfo[j].URL

	// content type & payload
	if t.With.PayloadURL != nil {
		b, err := getPayloadBytes(*t.With.PayloadURL)
		if err != nil {
			return nil, err
		} else {
			fo.Payload = b
		}
	}
	if t.With.PayloadJSON != nil {
		payloadBytes, err := json.Marshal(t.With.PayloadJSON)
		if err != nil {
			e := errors.New("unable to marshal JSON payload")
			log.Logger.WithStackTrace(err.Error()).Error(e)
			return nil, e
		} else {
			fo.Payload = payloadBytes
			fo.ContentType = contentTypeJSON
		}
	}

	return fo, nil
}

// resultForVersion collects Fortio result for a given version
func (t *collectHTTPTask) resultForVersion(j int) (*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio with proper options

	fo, err := t.getFortioOptions(j)
	if err != nil {
		return nil, err
	}
	log.Logger.Trace("got fortio options")
	log.Logger.Trace("URL: ", fo.URL)
	ifr, err := fhttp.RunHTTPTest(fo)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("fortio failed")
		if ifr == nil {
			log.Logger.Error("failed to get results since fortio run was aborted")
		}
	}
	log.Logger.Trace("ran fortio http test")
	return ifr, err
}

// Run executes this task
func (t *collectHTTPTask) Run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	if len(t.With.VersionInfo) == 0 {
		log.Logger.Error("collect task must specify info for at least one version")
		return errors.New("collect task must specify info for at least one version")
	}

	fm := make([]*fhttp.HTTPRunnerResults, len(t.With.VersionInfo))

	// run fortio queries for each version sequentially
	for j := range t.With.VersionInfo {
		log.Logger.Trace("initiating fortio for version ", j)
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

	// this task creates insights for different versions
	// initialize num versions
	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionInfo))
	if err != nil {
		return err
	}
	in := exp.Result.Insights

	for i := range t.With.VersionInfo {
		if fm[i] != nil {
			// request count
			m := iter8BuiltInPrefix + "/" + builtInHTTPRequestCountId
			mm := MetricMeta{
				Description: "number of requests sent",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, float64(fm[i].DurationHistogram.Count))

			// error count & rate
			val := float64(0)
			for code, count := range fm[i].RetCodes {
				if t.errorCode(code) {
					val += float64(count)
				}
			}
			// error count
			m = iter8BuiltInPrefix + "/" + builtInHTTPErrorCountId
			mm = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, val)

			// error-rate
			m = iter8BuiltInPrefix + "/" + builtInHTTPErrorRateId
			rc := float64(fm[i].DurationHistogram.Count)
			if rc != 0 {
				mm = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				in.updateMetric(m, mm, i, val/rc)
			}

			// mean-latency
			m = iter8BuiltInPrefix + "/" + builtInHTTPLatencyMeanId
			mm = MetricMeta{
				Description: "mean of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Avg)

			// stddev-latency
			m = iter8BuiltInPrefix + "/" + builtInHTTPLatencyStdDevId
			mm = MetricMeta{
				Description: "standard deviation of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.StdDev)

			// min-latency
			m = iter8BuiltInPrefix + "/" + builtInHTTPLatencyMinId
			mm = MetricMeta{
				Description: "minimum of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Min)

			// max-latency
			m = iter8BuiltInPrefix + "/" + builtInHTTPLatencyMaxId
			mm = MetricMeta{
				Description: "maximum of observed latency values",
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, i, 1000.0*fm[i].DurationHistogram.Max)

			// percentiles
			for _, p := range fm[i].DurationHistogram.Percentiles {
				m = fmt.Sprintf("%v/%v%v", iter8BuiltInPrefix, builtInHTTPLatencyPercentilePrefix, p.Percentile)
				mm = MetricMeta{
					Description: fmt.Sprintf("%v-th percentile of observed latency values", p.Percentile),
					Type:        GaugeMetricType,
					Units:       StringPointer("msec"),
				}
				in.updateMetric(m, mm, i, 1000.0*p.Value)
			}

			// latency histogram
			m = iter8BuiltInPrefix + "/" + builtInHTTPLatencyHistId
			mm = MetricMeta{
				Description: "Latency Histogram",
				Type:        HistogramMetricType,
				Units:       StringPointer("msec"),
			}
			lh := latencyHist(fm[i].DurationHistogram)
			in.updateMetric(m, mm, i, lh)
		}
	}
	return nil
}

// compute latency histogram by resampling
func latencyHist(hd *stats.HistogramData) []HistBucket {
	buckets := []HistBucket{}
	for _, v := range hd.Data {
		buckets = append(buckets, HistBucket{
			Lower: v.Start * 1000.0, // sec to msec
			Upper: v.End * 1000.0,
			Count: uint64(v.Count),
		})
	}
	return buckets
}
