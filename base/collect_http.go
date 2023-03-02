package base

import (
	"fmt"
	"io"
	"os"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	"fortio.org/fortio/stats"
	"github.com/imdario/mergo"
	log "github.com/iter8-tools/iter8/base/log"
)

// errorRange has lower and upper limits for HTTP status codes. HTTP status code within this range is considered an error
type errorRange struct {
	// Lower end of the range
	Lower *int `json:"lower,omitempty" yaml:"lower,omitempty"`
	// Upper end of the range
	Upper *int `json:"upper,omitempty" yaml:"upper,omitempty"`
}

// collectHTTPInputsHelper contains the inputs for one endpoint
type collectHTTPInputsHelper struct {
	// NumRequests is the number of requests to be sent to the app. Default value is 100.
	NumRequests *int64 `json:"numRequests,omitempty" yaml:"numRequests,omitempty"`
	// Duration of this task. Specified in the Go duration string format (example, 5s). If both duration and numRequests are specified, then duration is ignored.
	Duration *string `json:"duration,omitempty" yaml:"duration,omitempty"`
	// QPS is the number of requests per second sent to the app. Default value is 8.0.
	QPS *float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	// Connections is the number of number of parallel connections used to send load. Default value is 4.
	Connections *int `json:"connections,omitempty" yaml:"connections,omitempty"`
	// PayloadStr is the string data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to the app using this string as the payload.
	PayloadStr *string `json:"payloadStr,omitempty" yaml:"payloadStr,omitempty"`
	// PayloadFile is payload file. If this field is specified, Iter8 will send HTTP POST requests to the app using data in this file. If both `payloadStr` and `payloadFile` are specified, the former is ignored.
	PayloadFile *string `json:"payloadFile,omitempty" yaml:"payloadFile,omitempty"`
	// ContentType is the type of the payload. Indicated using the Content-Type HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to the app using this content type header value.
	ContentType *string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	// ErrorRanges is a list of errorRange values. Each range specifies an upper and/or lower limit on HTTP status codes. HTTP responses that fall within these error ranges are considered error. Default value is {{lower: 400},} - i.e., HTTP status codes >= 400 are considered as error.
	ErrorRanges []errorRange `json:"errorRanges,omitempty" yaml:"errorRanges,omitempty"`
	// Percentiles are the latency percentiles collected by this task. Percentile values have a single digit precision (i.e., rounded to one decimal place). Default value is {50.0, 75.0, 90.0, 95.0, 99.0, 99.9,}.
	Percentiles []float64 `json:"percentiles,omitempty" yaml:"percentiles,omitempty"`
	// HTTP headers to use in the query; optional
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// URL to use for querying the app
	URL string `json:"url" yaml:"url"`
	// AllowInitialErrors allows and doesn't abort on initial warmup errors
	AllowInitialErrors *bool `json:"allowInitialErrors,omitempty" yaml:"allowInitialErrors,omitempty"`
	// Warmup indicates if task execution is for warmup purposes; if so the results will be ignored
	Warmup *bool `json:"warmup,omitempty" yaml:"warmup,omitempty"`
}

// collectHTTPInputs contain the inputs to the metrics collection task to be executed.
type collectHTTPInputs struct {
	collectHTTPInputsHelper

	// Endpoints is used to define multiple endpoints to test
	Endpoints map[string]collectHTTPInputsHelper `json:"endpoints" yaml:"endpoints"`
}

const (
	// CollectHTTPTaskName is the name of this task which performs load generation and metrics collection.
	CollectHTTPTaskName = "http"
	// defaultQPS is the default number of queries per second
	defaultQPS = float32(8)
	// defaultHTTPNumRequests is the default number of queries (in total)
	defaultHTTPNumRequests = int64(100)
	// defaultHTTPConnections is the default number of connections (parallel go routines)
	defaultHTTPConnections = 4
	// httpMetricPrefix is the prefix for all metrics collected by this task
	httpMetricPrefix = "http"
	// the following are a list of names for metrics collected by this task
	builtInHTTPRequestCountID  = "request-count"
	builtInHTTPErrorCountID    = "error-count"
	builtInHTTPErrorRateID     = "error-rate"
	builtInHTTPLatencyMeanID   = "latency-mean"
	builtInHTTPLatencyStdDevID = "latency-stddev"
	builtInHTTPLatencyMinID    = "latency-min"
	builtInHTTPLatencyMaxID    = "latency-max"
	builtInHTTPLatencyHistID   = "latency"
	// prefix used in latency percentile metric names
	// example: latency-p75.0 is the 75th percentile latency
	builtInHTTPLatencyPercentilePrefix = "latency-p"
)

var (
	// defaultErrorRanges is set so that status codes 400 and above are considered errors
	defaultErrorRanges = []errorRange{{Lower: intPointer(400)}}
	// defaultPercentiles are the default latency percentiles computed by this task
	defaultPercentiles = [...]float64{50.0, 75.0, 90.0, 95.0, 99.0, 99.9}
)

// errorCode checks if a given code is an error code
func (t *collectHTTPTask) errorCode(code int) bool {
	// connection failure
	if code == -1 {
		return true
	}
	// HTTP errors
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

// collectHTTPTask enables load testing of HTTP services.
type collectHTTPTask struct {
	// TaskMeta has fields common to all tasks
	TaskMeta
	// With contains the inputs to this task
	With collectHTTPInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectHTTPTask) initializeDefaults() {
	if t.With.NumRequests == nil && t.With.Duration == nil {
		t.With.NumRequests = int64Pointer(defaultHTTPNumRequests)
	}
	if t.With.QPS == nil {
		t.With.QPS = float32Pointer(defaultQPS)
	}
	if t.With.Connections == nil {
		t.With.Connections = intPointer(defaultHTTPConnections)
	}
	if t.With.ErrorRanges == nil {
		t.With.ErrorRanges = defaultErrorRanges
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
	if t.With.AllowInitialErrors == nil {
		t.With.AllowInitialErrors = BoolPointer(false)
	}
}

// validateInputs for this task
func (t *collectHTTPTask) validateInputs() error {
	return nil
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func getFortioOptions(c collectHTTPInputsHelper) (*fhttp.HTTPRunnerOptions, error) {
	// basic runner
	fo := &fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			RunType:     "Iter8 load test",
			QPS:         float64(*c.QPS),
			NumThreads:  *c.Connections,
			Percentiles: c.Percentiles,
			Out:         io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: c.URL,
		},
		AllowInitialErrors: *c.AllowInitialErrors,
	}

	// num requests
	if c.NumRequests != nil {
		fo.RunnerOptions.Exactly = *c.NumRequests
	}

	// add duration
	var duration time.Duration
	var err error
	if c.Duration != nil {
		duration, err = time.ParseDuration(*c.Duration)
		if err == nil {
			fo.RunnerOptions.Duration = duration
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("unable to parse duration")
			return nil, err
		}
	}

	// content type & payload
	if c.ContentType != nil {
		fo.ContentType = *c.ContentType
	}
	if c.PayloadStr != nil {
		fo.Payload = []byte(*c.PayloadStr)
	}
	if c.PayloadFile != nil {
		b, err := os.ReadFile(*c.PayloadFile)
		if err != nil {
			return nil, err
		}
		fo.Payload = b
	}

	// headers
	for key, value := range c.Headers {
		if err = fo.AddAndValidateExtraHeader(key + ":" + value); err != nil {
			log.Logger.WithStackTrace("unable to add header").Error(err)
			return nil, err
		}
	}

	return fo, nil
}

// getFortioResults collects Fortio run results
// func (t *collectHTTPTask) getFortioResults() (*fhttp.HTTPRunnerResults, error) {
// key is the metric prefix
func (t *collectHTTPTask) getFortioResults() (map[string]*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio with proper options

	var err error
	results := map[string]*fhttp.HTTPRunnerResults{}
	if len(t.With.Endpoints) > 0 {
		log.Logger.Trace("multiple endpoints")
		for endpointID, endpoint := range t.With.Endpoints {
			endpoint := endpoint // prevent implicit memory aliasing
			log.Logger.Trace(fmt.Sprintf("endpoint: %s", endpointID))

			// merge endpoint config with baseline config
			if err := mergo.Merge(&endpoint, t.With.collectHTTPInputsHelper); err != nil {
				log.Logger.Error(fmt.Sprintf("could not merge Fortio options for endpoint \"%s\"", endpointID))
				return nil, err
			}

			efo, err := getFortioOptions(endpoint)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("could not get Fortio options for endpoint \"%s\"", endpointID))
				return nil, err
			}

			log.Logger.Trace("got fortio options")
			log.Logger.Trace("URL: ", efo.URL)

			ifr, err := fhttp.RunHTTPTest(efo)
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("fortio failed")
				if ifr == nil {
					log.Logger.Error("failed to get results since fortio run was aborted")
				}
			}

			log.Logger.Trace("ran fortio http test")

			results[httpMetricPrefix+"-"+endpointID] = ifr
		}
	} else {
		fo, err := getFortioOptions(t.With.collectHTTPInputsHelper)
		if err != nil {
			log.Logger.Error("could not get Fortio options")
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

		results[httpMetricPrefix] = ifr
	}

	return results, err
}

// run executes this task
func (t *collectHTTPTask) run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	// run fortio
	data, err := t.getFortioResults()
	if err != nil {
		return err
	}

	// ignore results if warmup
	if t.With.Warmup != nil && *t.With.Warmup {
		log.Logger.Debug("warmup: ignoring results")
		return nil
	}

	// this task populates insights in the experiment
	// hence, initialize insights with num versions (= 1)
	err = exp.Result.initInsightsWithNumVersions(1)
	if err != nil {
		return err
	}
	in := exp.Result.Insights

	for provider, data := range data {
		// request count
		m := provider + "/" + builtInHTTPRequestCountID
		mm := MetricMeta{
			Description: "number of requests sent",
			Type:        CounterMetricType,
		}
		if err = in.updateMetric(m, mm, 0, float64(data.DurationHistogram.Count)); err != nil {
			return err
		}

		// error count & rate
		val := float64(0)
		for code, count := range data.RetCodes {
			if t.errorCode(code) {
				val += float64(count)
			}
		}
		// error count
		m = provider + "/" + builtInHTTPErrorCountID
		mm = MetricMeta{
			Description: "number of responses that were errors",
			Type:        CounterMetricType,
		}
		if err = in.updateMetric(m, mm, 0, val); err != nil {
			return err
		}

		// error-rate
		m = provider + "/" + builtInHTTPErrorRateID
		rc := float64(data.DurationHistogram.Count)
		if rc != 0 {
			mm = MetricMeta{
				Description: "fraction of responses that were errors",
				Type:        GaugeMetricType,
			}
			if err = in.updateMetric(m, mm, 0, val/rc); err != nil {
				return err
			}
		}

		// mean-latency
		m = provider + "/" + builtInHTTPLatencyMeanID
		mm = MetricMeta{
			Description: "mean of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		if err = in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Avg); err != nil {
			return err
		}

		// stddev-latency
		m = provider + "/" + builtInHTTPLatencyStdDevID
		mm = MetricMeta{
			Description: "standard deviation of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		if err = in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.StdDev); err != nil {
			return err
		}

		// min-latency
		m = provider + "/" + builtInHTTPLatencyMinID
		mm = MetricMeta{
			Description: "minimum of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		if err = in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Min); err != nil {
			return err
		}

		// max-latency
		m = provider + "/" + builtInHTTPLatencyMaxID
		mm = MetricMeta{
			Description: "maximum of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		if err = in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Max); err != nil {
			return err
		}

		// percentiles
		for _, p := range data.DurationHistogram.Percentiles {
			m = fmt.Sprintf("%v/%v%v", provider, builtInHTTPLatencyPercentilePrefix, p.Percentile)
			mm = MetricMeta{
				Description: fmt.Sprintf("%v-th percentile of observed latency values", p.Percentile),
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			if err = in.updateMetric(m, mm, 0, 1000.0*p.Value); err != nil {
				return err
			}
		}

		// latency histogram
		m = httpMetricPrefix + "/" + builtInHTTPLatencyHistID
		mm = MetricMeta{
			Description: "Latency Histogram",
			Type:        HistogramMetricType,
			Units:       StringPointer("msec"),
		}
		lh := latencyHist(data.DurationHistogram)
		if err = in.updateMetric(m, mm, 0, lh); err != nil {
			return err
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
