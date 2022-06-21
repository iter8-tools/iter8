package base

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"fortio.org/fortio/fhttp"
	fortioLog "fortio.org/fortio/log"
	"fortio.org/fortio/periodic"
	"fortio.org/fortio/stats"
	log "github.com/iter8-tools/iter8/base/log"
)

// errorRange has lower and upper limits for HTTP status codes. HTTP status code within this range is considered an error
type errorRange struct {
	// Lower end of the range
	Lower *int `json:"lower,omitempty" yaml:"lower,omitempty"`
	// Upper end of the range
	Upper *int `json:"upper,omitempty" yaml:"upper,omitempty"`
}

// collectHTTPInputs contain the inputs to the metrics collection task to be executed.
type collectHTTPInputs struct {
	// NumRequests is the number of requests to be sent to the app. Default value is 100.
	NumRequests *int64 `json:"numRequests,omitempty" yaml:"numRequests,omitempty"`
	// Duration of this task. Specified in the Go duration string format (example, 5s). If both duration and numQueries are specified, then duration is ignored.
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
	builtInHTTPRequestCountId  = "request-count"
	builtInHTTPErrorCountId    = "error-count"
	builtInHTTPErrorRateId     = "error-rate"
	builtInHTTPLatencyMeanId   = "latency-mean"
	builtInHTTPLatencyStdDevId = "latency-stddev"
	builtInHTTPLatencyMinId    = "latency-min"
	builtInHTTPLatencyMaxId    = "latency-max"
	builtInHTTPLatencyHistId   = "latency"
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
}

// validateInputs for this task
func (t *collectHTTPTask) validateInputs() error {
	return nil
}

// getFortioOptions constructs Fortio's HTTP runner options based on collect task inputs
func (t *collectHTTPTask) getFortioOptions() (*fhttp.HTTPRunnerOptions, error) {
	fortioLog.SetOutput(io.Discard)
	// basic runner
	fo := &fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			RunType:     "Iter8 load test",
			QPS:         float64(*t.With.QPS),
			NumThreads:  *t.With.Connections,
			Percentiles: t.With.Percentiles,
			Out:         io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: t.With.URL,
		},
	}

	// num requests
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
	if t.With.PayloadFile != nil {
		b, err := ioutil.ReadFile(*t.With.PayloadFile)
		if err != nil {
			return nil, err
		} else {
			fo.Payload = b
		}
	}

	// headers
	for key, value := range t.With.Headers {
		fo.AddAndValidateExtraHeader(key + ":" + value)
	}

	return fo, nil
}

// getFortioResults collects Fortio run results
func (t *collectHTTPTask) getFortioResults() (*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio with proper options

	fo, err := t.getFortioOptions()
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

	// this task populates insights in the experiment
	// hence, initialize insights with num versions (= 1)
	err = exp.Result.initInsightsWithNumVersions(1)
	if err != nil {
		return err
	}
	in := exp.Result.Insights

	if data != nil {
		// request count
		m := httpMetricPrefix + "/" + builtInHTTPRequestCountId
		mm := MetricMeta{
			Description: "number of requests sent",
			Type:        CounterMetricType,
		}
		in.updateMetric(m, mm, 0, float64(data.DurationHistogram.Count))

		// error count & rate
		val := float64(0)
		for code, count := range data.RetCodes {
			if t.errorCode(code) {
				val += float64(count)
			}
		}
		// error count
		m = httpMetricPrefix + "/" + builtInHTTPErrorCountId
		mm = MetricMeta{
			Description: "number of responses that were errors",
			Type:        CounterMetricType,
		}
		in.updateMetric(m, mm, 0, val)

		// error-rate
		m = httpMetricPrefix + "/" + builtInHTTPErrorRateId
		rc := float64(data.DurationHistogram.Count)
		if rc != 0 {
			mm = MetricMeta{
				Description: "fraction of responses that were errors",
				Type:        GaugeMetricType,
			}
			in.updateMetric(m, mm, 0, val/rc)
		}

		// mean-latency
		m = httpMetricPrefix + "/" + builtInHTTPLatencyMeanId
		mm = MetricMeta{
			Description: "mean of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Avg)

		// stddev-latency
		m = httpMetricPrefix + "/" + builtInHTTPLatencyStdDevId
		mm = MetricMeta{
			Description: "standard deviation of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.StdDev)

		// min-latency
		m = httpMetricPrefix + "/" + builtInHTTPLatencyMinId
		mm = MetricMeta{
			Description: "minimum of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Min)

		// max-latency
		m = httpMetricPrefix + "/" + builtInHTTPLatencyMaxId
		mm = MetricMeta{
			Description: "maximum of observed latency values",
			Type:        GaugeMetricType,
			Units:       StringPointer("msec"),
		}
		in.updateMetric(m, mm, 0, 1000.0*data.DurationHistogram.Max)

		// percentiles
		for _, p := range data.DurationHistogram.Percentiles {
			m = fmt.Sprintf("%v/%v%v", httpMetricPrefix, builtInHTTPLatencyPercentilePrefix, p.Percentile)
			mm = MetricMeta{
				Description: fmt.Sprintf("%v-th percentile of observed latency values", p.Percentile),
				Type:        GaugeMetricType,
				Units:       StringPointer("msec"),
			}
			in.updateMetric(m, mm, 0, 1000.0*p.Value)
		}

		// latency histogram
		m = httpMetricPrefix + "/" + builtInHTTPLatencyHistId
		mm = MetricMeta{
			Description: "Latency Histogram",
			Type:        HistogramMetricType,
			Units:       StringPointer("msec"),
		}
		lh := latencyHist(data.DurationHistogram)
		in.updateMetric(m, mm, 0, lh)
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
