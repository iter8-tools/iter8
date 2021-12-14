package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"

	"fortio.org/fortio/fhttp"
	fortioLog "fortio.org/fortio/log"
	"fortio.org/fortio/periodic"
	"fortio.org/fortio/stats"
	"github.com/go-playground/validator"
	log "github.com/iter8-tools/iter8/base/log"
)

// version contains header and url information needed to send requests to each version.
type version struct {
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers" yaml:"headers"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url" validate:"required,url"`
}

// HTTP status code within this range is considered an error
type errorRange struct {
	Lower *int `json:"lower" yaml:"lower" validate:"required_without=Upper"`
	Upper *int `json:"upper" yaml:"upper" validate:"required_without=Lower"`
}

// collectInputs contain the inputs to the metrics collection task to be executed.
type collectInputs struct {
	// NumRequests is the number of requests to be sent to each version. Default value is 100.
	NumRequests *int64 `json:"numRequests" yaml:"numRequests"`
	// Duration of the metrics/collect task run. Specified in the Go duration string format (example, 5s). If both duration and numQueries are specified, then duration is ignored.
	Duration *string `json:"duration" yaml:"duration"`
	// QPS is the number of queries per second sent to each version. Default value is 8.0.
	QPS *float32 `json:"qps" yaml:"qps"`
	// Connections is the number of number of parallel connections used to send load. Default value is 4.
	Connections *int `json:"connections" yaml:"connections"`
	// PayloadStr is the string data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to versions using this string as the payload.
	PayloadStr *string `json:"payloadStr" yaml:"payloadStr"`
	// PayloadURL is the URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions using data downloaded from this URL as the payload. If both `payloadStr` and `payloadURL` is specified, the former is ignored.
	PayloadURL *string `json:"payloadURL" yaml:"payloadURL"`
	// ContentType is the type of the payload. Indicated using the Content-Type HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to versions using this content type header value.
	ContentType *string `json:"contentType" yaml:"contentType" validate:"required_with=PayloadURL"`
	// ErrorRanges is a list of errorRange values. Each range specifies an upper and/or lower limit on HTTP status codes. HTTP responses that fall within these error ranges are considered error. Default value is {{lower: 400},} - i.e., HTTP status codes >= 400 are considered as error.
	ErrorRanges []errorRange `json:"errorRanges" yaml:"errorRanges"`
	// Percentiles are the latency percentiles computed by this task. Percentile values have a single digit precision (i.e., rounded to one decimal place). Default value is {50.0, 75.0, 90.0, 95.0, 99.0, 99.9,}.
	Percentiles []float64 `json:"percentiles" yaml:"percentiles" validate:"unique,dive,gte=0.0,lte=100.0"`
	// A non-empty list of version values.
	VersionInfo []*version `json:"versionInfo" yaml:"versionInfo" validate:"required,notallnil"`
}

const (
	CollectTaskName            = "gen-load-and-collect-metrics"
	defaultQPS                 = float32(8)
	defaultNumRequests         = int64(100)
	defaultConnections         = 4
	requestCountMetricName     = "request-count"
	errorCountMetricName       = "error-count"
	errorRateMetricName        = "error-rate"
	meanLatencyMetricName      = "mean-latency"
	stdDevMetricName           = "stddev-latency"
	minLatencyMetricName       = "min-latency"
	maxLatencyMetricName       = "max-latency"
	latencyHistogramMetricName = "latency"
)

var (
	defaultErrorRanges = []errorRange{{Lower: intPointer(400)}}
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
	With collectInputs `json:"with" yaml:"with" validate:"required"`
}

// notAllNil validates that not all entries in a slice are nil
func notAllNil(fl validator.FieldLevel) bool {
	slice, ok := fl.Field().Interface().([]*version)
	if !ok {
		log.Logger.WithStackTrace(fl.Field().String()).Error("unable to convert field to slice of version pointers")
		return false
	}
	allnil := true
	for _, s := range slice {
		if s != nil {
			allnil = false
			break
		}
	}
	return !allnil
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
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("invalid collect task specification")
			return nil, err
		}

		validate := validator.New()
		// returns nil or ValidationErrors ( []FieldError )
		validate.RegisterValidation("notallnil", notAllNil)

		err = validate.Struct(ct)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("invalid collect task specification")
			return nil, err
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
	if t.With.Connections == nil {
		t.With.Connections = intPointer(defaultConnections)
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
			NumThreads:  *t.With.Connections,
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

// GetName returns the name of the assess task
func (t *collectTask) GetName() string {
	return CollectTaskName
}

// Run executes the metrics/collect task
func (t *collectTask) Run(exp *Experiment) error {
	var err error
	t.initializeDefaults()

	if len(t.With.VersionInfo) == 0 {
		log.Logger.Error("collect task must specify info for at least one version")
		return errors.New("collect task must specify info for at least one version")
	}

	fm := make([]*fhttp.HTTPRunnerResults, len(t.With.VersionInfo))

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

	// initialize insights if required
	if exp.Result.Insights == nil {
		exp.Result.InitInsights(len(t.With.VersionInfo), []InsightType{InsightTypeMetrics})
	}

	in := exp.Result.Insights

	// set hist metrics insight type (if needed)
	in.setInsightType(InsightTypeHistMetrics)

	// initialize metric values (if needed)
	err = in.initMetricValues(len(t.With.VersionInfo))
	if err != nil {
		return err
	}

	// compute min latency, max latency and bucket sizes needed for latency histograms
	m := iter8BuiltInPrefix + "/" + latencyHistogramMetricName
	var xMin float64 = 0   // msec
	var xMax float64 = 200 // msec
	if mi, ok := in.MetricsInfo[m]; ok {
		xMin = *mi.XMin
		xMax = *mi.XMax
	}
	xMin = aggregateLatencyBound(fm, xMin, true)
	xMax = aggregateLatencyBound(fm, xMax, false)

	for i := range t.With.VersionInfo {
		if fm[i] != nil {
			// request count
			m := iter8BuiltInPrefix + "/" + requestCountMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "number of requests sent",
				Type:        CounterMetricType,
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], float64(fm[i].DurationHistogram.Count))

			// error count & rate
			val := float64(0)
			for code, count := range fm[i].RetCodes {
				if t.errorCode(code) {
					val += float64(count)
				}
			}
			// error count
			m = iter8BuiltInPrefix + "/" + errorCountMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], val)

			// error-rate
			m = iter8BuiltInPrefix + "/" + errorRateMetricName
			rc := float64(fm[i].DurationHistogram.Count)
			if rc != 0 {
				in.MetricsInfo[m] = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				in.MetricValues[i][m] = append(in.MetricValues[i][m], val/rc)
			}

			// mean-latency
			m = iter8BuiltInPrefix + "/" + meanLatencyMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "mean of observed latency values",
				Type:        GaugeMetricType,
				Units:       stringPointer("msec"),
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], 1000.0*fm[i].DurationHistogram.Avg)

			// stddev-latency
			m = iter8BuiltInPrefix + "/" + stdDevMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "standard deviation of observed latency values",
				Type:        GaugeMetricType,
				Units:       stringPointer("msec"),
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], 1000.0*fm[i].DurationHistogram.StdDev)

			// min-latency
			m = iter8BuiltInPrefix + "/" + minLatencyMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "minimum of observed latency values",
				Type:        GaugeMetricType,
				Units:       stringPointer("msec"),
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], 1000.0*fm[i].DurationHistogram.Min)

			// max-latency
			m = iter8BuiltInPrefix + "/" + maxLatencyMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "maximum of observed latency values",
				Type:        GaugeMetricType,
				Units:       stringPointer("msec"),
			}
			in.MetricValues[i][m] = append(in.MetricValues[i][m], 1000.0*fm[i].DurationHistogram.Max)

			// percentiles
			for _, p := range fm[i].DurationHistogram.Percentiles {
				m = iter8BuiltInPrefix + "/" + fmt.Sprintf("p%0.1f", p.Percentile)
				in.MetricsInfo[m] = MetricMeta{
					Description: fmt.Sprintf("%0.1f-th percentile of observed latency values", p.Percentile),
					Type:        GaugeMetricType,
					Units:       stringPointer("msec"),
				}
				in.MetricValues[i][m] = append(in.MetricValues[i][m], 1000.0*p.Value)
			}

			// latency histogram
			m = iter8BuiltInPrefix + "/" + latencyHistogramMetricName
			in.MetricsInfo[m] = MetricMeta{
				Description: "Latency Histogram",
				Type:        HistogramMetricType,
				Units:       stringPointer("msec"),
				XMin:        float64Pointer(xMin),
				XMax:        float64Pointer(xMax),
				NumBuckets:  intPointer(20),
			}
			lh := latencyHist(in.MetricsInfo[m], fm[i].DurationHistogram)
			if in.MetricValues[i][m] == nil {
				in.MetricValues[i][m] = lh
			} else {
				in.MetricValues[i][m], err = vectorSum(in.MetricValues[i][m], lh)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// sum up two vectors
func vectorSum(a []float64, b []float64) ([]float64, error) {
	if len(a) != len(b) {
		log.Logger.Error("vector lengths do not match: ", len(a), " != ", len(b))
		return nil, fmt.Errorf("vector lengths do not match: %v != %v", len(a), len(b))
	}
	sum := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		sum[i] = a[i] + b[i]
	}
	return sum, nil
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

// aggregateLatencyBound returns min or max latency value
func aggregateLatencyBound(fm []*fhttp.HTTPRunnerResults, val float64, min bool) float64 {
	extremeVal := val
	for i := 0; i < len(fm); i++ {
		if fm[i] != nil {
			if fm[i].DurationHistogram != nil {
				if min {
					extremeVal = math.Min(extremeVal, 1000.0*fm[i].DurationHistogram.Min)
				} else {
					extremeVal = math.Max(extremeVal, 1000.0*fm[i].DurationHistogram.Max)
				}
			}
		}
	}
	return extremeVal
}

// compute latency histogram by resampling
func latencyHist(mm MetricMeta, hd *stats.HistogramData) []float64 {
	rand.Seed(0) // we will make the random number generation deterministic
	// initialize histogram
	hist := make([]float64, *mm.NumBuckets)
	if hd == nil {
		return hist
	}

	// create sample
	var sample []float64
	for i := 0; i < len(hd.Data); i++ {
		bucket := hd.Data[i]
		for j := 0; j < int(bucket.Count); j++ {
			// get a point; convert to msec
			point := 1000.0 * (bucket.Start + (bucket.End-bucket.Start)*rand.Float64())
			sample = append(sample, point)
		}
	}

	// update hist
	for i := 0; i < len(sample); i++ {
		width := (*mm.XMax - *mm.XMin) / float64(*mm.NumBuckets)
		ind := int((sample[i] - (*mm.XMin)) / width)
		if ind < 0 {
			log.Logger.Warn("sample index out of range: ", ind)
			ind = 0
		}
		if ind >= len(hist) {
			log.Logger.Warn("sample index out of range: ", ind)
			ind = len(hist) - 1
		}
		hist[ind]++
	}
	return hist
}
