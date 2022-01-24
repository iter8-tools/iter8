package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/iter8-tools/iter8/base/log"
	"github.com/montanaflynn/stats"
)

// Task is an object that can be run
type Task interface {
	// validate inputs for this task
	validateInputs() error
	// initializeDefault values for inputs to this task
	initializeDefaults()
	// Run this task
	Run(exp *Experiment) error
}

// ExperimentSpec is the experiment spec
type ExperimentSpec []Task

// Experiment specification and result
type Experiment struct {
	// Tasks is the sequence of tasks that constitute this experiment
	Tasks ExperimentSpec
	// Result is the current results from this experiment.
	// The experiment may not have completed in which case results may be partial.
	Result *ExperimentResult
}

// UnmarshallJSON will unmarshal an experiment spec from bytes
func (s ExperimentSpec) UnmarshalJSON(data []byte) error {
	var v []taskMeta
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	for _, t := range v {
		if (t.Task == nil || len(*t.Task) == 0) && (t.Run == nil) {
			err := fmt.Errorf("invalid task found without a task name or a run command")
			log.Logger.Error(err)
			return err
		}

		// get byte data for this task
		tBytes, _ := json.Marshal(t)
		var tsk Task
		// this is a run task
		if t.Run != nil {
			tsk = &runTask{}
		} else {
			// this is some other task
			switch *t.Task {
			case CollectTaskName:
				tsk = &collectTask{}
			case CollectGPRCTaskName:
				tsk = &collectGRPCTask{}
			case AssessTaskName:
				tsk = &assessTask{}
			default:
				log.Logger.Error("unknown task: " + *t.Task)
				return errors.New("unknown task: " + *t.Task)
			}
			json.Unmarshal(tBytes, tsk)
			s = append(s, tsk)
		}
	}
	log.Logger.Trace("constructed experiment spec of length: ", len(s))
	return nil
}

// GetIf returns the condition (if any) which determine
// whether of not if this task needs to run
func GetIf(t Task) *string {
	var jsonBytes []byte
	var tm taskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to taskMeta
	_ = json.Unmarshal(jsonBytes, &tm)
	return tm.If
}

// GetName returns the name of this task
func GetName(t Task) *string {
	var jsonBytes []byte
	var tm taskMeta
	// convert t to jsonBytes
	jsonBytes, _ = json.Marshal(t)
	// convert jsonBytes to taskMeta
	_ = json.Unmarshal(jsonBytes, &tm)

	if tm.Task == nil {
		if tm.Run != nil {
			return StringPointer(RunTaskName)
		}
	} else {
		return tm.Task
	}
	log.Logger.Error("Task specification with no name or run value")
	return nil
}

// ExperimentResult defines the current results from the experiment
type ExperimentResult struct {
	// StartTime is the time when the experiment run started
	StartTime time.Time `json:"startTime" yaml:"startTime"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"numCompletedTasks" yaml:"numCompletedTasks"`

	// Failure is true if any of its tasks failed
	Failure bool `json:"failure" yaml:"failure"`

	// Insights produced in this experiment
	Insights *Insights `json:"insights,omitempty" yaml:"insights,omitempty"`
}

// Insights records the number of versions in this experiment,
// metric values and SLO indicators for each version,
// metrics metadata for all metrics, and
// SLO definitions for all SLOs
type Insights struct {
	// NumVersions is the number of app versions detected by Iter8
	NumVersions int `json:"numVersions" yaml:"numVersions"`

	// MetricsInfo identifies the metrics involved in this experiment
	MetricsInfo map[string]MetricMeta `json:"metricsInfo,omitempty" yaml:"metricsInfo,omitempty"`

	// MetricValues:
	// the outer slice must be the same length as the number of app versions
	// the map key must match name of a metric in MetricsInfo
	// the inner slice contains the list of all observed metric values for given version and given metric; float value [i]["foo/bar"][k] is the [k]th observation for version [i] for the metric bar under backend foo.
	MetricValues []map[string][]float64 `json:"metricValues,omitempty" yaml:"metricValues,omitempty"`

	// HistMetricValues:
	// the outer slice must be the same length as the number of app versions
	// the map key must match name of a histogram metric in MetricsInfo
	// the inner slice contains the list of all observed histogram buckets for a given version and given metric; value [i]["foo/bar"][k] is the [k]th observed bucket for version [i] for the hist metric `bar` under backend `foo`.
	HistMetricValues []map[string][]HistBucket `json:"histMetricValues,omitempty" yaml:"histMetricValues,omitempty"`

	// SLOs involved in this experiment
	SLOs []SLO `json:"SLOs,omitempty" yaml:"SLOs,omitempty"`

	// SLOsSatisfied:
	// the outer slice must be of the same length as SLOs
	// the length of the inner slice must be the number of app versions
	// the boolean value at [i][j] indicate if SLO [i] is satisfied by version [j]
	SLOsSatisfied [][]bool `json:"SLOsSatisfied,omitempty" yaml:"SLOsSatisfied,omitempty"`
}

// MetricMeta describes a metric
type MetricMeta struct {
	Description string     `json:"description" yaml:"description"`
	Units       *string    `json:"units,omitempty" yaml:"units,omitempty"`
	Type        MetricType `json:"type" yaml:"type"`
}

// SLO is a service level objective
type SLO struct {
	// Metric is the fully qualified metric name (i.e., in the backendName/metricName format)
	Metric string `json:"metric" yaml:"metric"`

	// UpperLimit is the maximum acceptable value of the metric.
	UpperLimit *float64 `json:"upperLimit,omitempty" yaml:"upperLimit,omitempty"`

	// LowerLimit is the minimum acceptable value of the metric.
	LowerLimit *float64 `json:"lowerLimit,omitempty" yaml:"lowerLimit,omitempty"`
}

type taskMeta struct {
	// Task is the name of the task
	Task *string `json:"task,omitempty" yaml:"task,omitempty"`
	// Run is the script used in a run task
	// Specify either Task or Run but not both
	Run *string `json:"run,omitempty" yaml:"run,omitempty"`
	// If is the condition used to determine if this task needs to run.
	If *string `json:"if,omitempty" yaml:"if,omitempty"`
}

// metricTypeMatch checks if metric value is a match for its type
func metricTypeMatch(t MetricType, val interface{}) bool {
	switch v := val.(type) {
	case float64:
		if t == CounterMetricType || t == GaugeMetricType {
			return true
		} else {
			return false
		}
	case []float64:
		if t == SampleMetricType || t == HistogramMetricType {
			return true
		} else {
			return false
		}
	default:
		log.Logger.Error("unsupported type for metric value: ", v)
		return false
	}
}

// updateMetricValue update a metric value for a given version
func (in *Insights) updateMetricValue(m string, i int, val interface{}) error {
	switch v := val.(type) {
	case float64:
		in.MetricValues[i][m] = append(in.MetricValues[i][m], val.(float64))
		return nil
	case []float64:
		in.MetricValues[i][m] = append(in.MetricValues[i][m], val.([]float64)...)
		return nil
	default:
		err := fmt.Errorf("unsupported type for metric value: %s", v)
		log.Logger.Error(err)
		return err
	}
}

// registerMetric registers a new metric by adding its meta data
func (in *Insights) registerMetric(m string, mm MetricMeta) error {
	if old, ok := in.MetricsInfo[m]; ok && !reflect.DeepEqual(old, mm) {
		err := fmt.Errorf("old and new metric meta for %v differ", m)
		log.Logger.WithStackTrace(fmt.Sprintf("old: %v \nnew: %v", old, mm)).Error(err)
		return err
	}
	in.MetricsInfo[m] = mm
	return nil
}

// updateMetric registers a metric and adds a metric value for a given version
func (in *Insights) updateMetric(m string, mm MetricMeta, i int, val interface{}) error {
	var err error
	if !metricTypeMatch(mm.Type, val) {
		err = fmt.Errorf("metric value and type are incompatible; name: %v meta: %v version: %v value: %v", m, mm, i, val)
		log.Logger.Error(err)
		return err
	}

	if in.NumVersions <= i {
		err := fmt.Errorf("insufficient number of versions %v with version index %v", in.NumVersions, i)
		log.Logger.Error(err)
		return err
	}

	err = in.registerMetric(m, mm)
	if err != nil {
		return err
	}

	// update metric value
	return in.updateMetricValue(m, i, val)
}

// setSLOs sets the SLOs field in insights
// if this function is called multiple times (example, due to looping), then
// it is intended to be called with the same argument each time
func (in *Insights) setSLOs(slos []SLO) error {
	if in.SLOs != nil {
		if reflect.DeepEqual(in.SLOs, slos) {
			return nil
		} else {
			e := fmt.Errorf("old and new value of slos conflict")
			log.Logger.WithStackTrace(fmt.Sprint("old: ", in.SLOs, "new: ", slos)).Error(e)
			return e
		}
	}
	// LHS will be nil
	in.SLOs = slos
	return nil
}

// initializeSLOsSatisfied initializes the SLOs satisfied field
func (e *Experiment) initializeSLOsSatisfied() error {
	if e.Result.Insights.SLOsSatisfied != nil {
		return nil // already initialized
	}
	// LHS will be nil
	e.Result.Insights.SLOsSatisfied = make([][]bool, len(e.Result.Insights.SLOs))
	for i := 0; i < len(e.Result.Insights.SLOs); i++ {
		e.Result.Insights.SLOsSatisfied[i] = make([]bool, e.Result.Insights.NumVersions)
	}
	return nil
}

func (e *Experiment) InitResults() {
	e.Result = &ExperimentResult{
		StartTime:         time.Now(),
		NumCompletedTasks: 0,
		Failure:           false,
	}
}

// initInsightsWithNumVersions is also going to initialize insights data structure
// insights data structure contains metrics data structures, so this will also
// init metrics
func (r *ExperimentResult) initInsightsWithNumVersions(n int) error {
	if r.Insights != nil {
		if r.Insights.NumVersions != n {
			err := fmt.Errorf("inconsistent number for app versions; old (%v); new (%v)", r.Insights.NumVersions, n)
			log.Logger.Error(err)
			return err
		}
	} else {
		r.Insights = &Insights{
			NumVersions: n,
		}
	}
	r.Insights.initMetrics()
	return nil
}

// initMetrics initializes the data structes inside insights that will hold metrics
func (in *Insights) initMetrics() error {
	if len(in.MetricValues) != len(in.MetricsInfo) {
		err := fmt.Errorf("inconsistent number for app versions in metric values (%v), metric info (%v)", len(in.MetricValues), len(in.MetricsInfo))
		log.Logger.Error(err)
		return err
	}
	if in.MetricValues != nil {
		if len(in.MetricValues) != in.NumVersions {
			err := fmt.Errorf("inconsistent number for app versions in metric values (%v), num versions (%v)", len(in.MetricValues), in.NumVersions)
			log.Logger.Error(err)
			return err
		} else {
			return nil
		}
	}
	// at this point, there are no known metrics,
	// but there are in.NumVersions versions
	// Initialize metrics info
	in.MetricsInfo = make(map[string]MetricMeta)
	// Initialize metric values for each version
	in.MetricValues = make([]map[string][]float64, in.NumVersions)
	for i := 0; i < in.NumVersions; i++ {
		in.MetricValues[i] = make(map[string][]float64)
	}
	return nil
}

// getMetricFromValuesMap gets the value of the given counter or gauge metric, for the given version, from metric values map
func (in *Insights) getMetricFromValuesMap(i int, m string) *float64 {
	if mm, ok := in.MetricsInfo[m]; ok {
		log.Logger.Tracef("found metric info for %v", m)
		if (mm.Type != CounterMetricType) && (mm.Type != GaugeMetricType) {
			log.Logger.Errorf("metric %v is not of type counter or gauge", m)
			return nil
		}
		l := len(in.MetricValues)
		if l <= i {
			log.Logger.Warnf("metric values not found for version %v; initialized for %v versions", i, l)
			return nil
		}
		log.Logger.Tracef("metric values found for version %v", i)
		// grab the metric value and return
		if vals, ok := in.MetricValues[i][m]; ok {
			log.Logger.Tracef("found metric value for version %v and metric %v", i, m)
			if len(vals) > 0 {
				return float64Pointer(vals[len(vals)-1])
			}
		}
		log.Logger.Infof("could not find metric value for version %v and metric %v", i, m)
	}
	log.Logger.Infof("could not find metric info for %v", m)
	return nil
}

// getSampleAggregation aggregates the given base metric for the given version (i) with the given aggregation (a)
func (in *Insights) getSampleAggregation(i int, baseMetric string, a string) *float64 {
	at := AggregationType(a)
	vals := in.MetricValues[i][baseMetric]
	if len(vals) == 0 {
		log.Logger.Infof("metric %v for version %v has no sample", baseMetric, i)
	}
	if len(vals) == 1 {
		log.Logger.Warnf("metric %v for version %v has sample of size 1", baseMetric, i)
		return float64Pointer(vals[0])
	}
	switch at {
	case MeanAggregator:
		agg, err := stats.Mean(vals)
		if err == nil {
			return float64Pointer(agg)
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("aggregation error for version %v, metric %v, and aggregation func %v", i, baseMetric, a)
			return nil
		}
	case StdDevAggregator:
		agg, err := stats.StandardDeviation(vals)
		if err == nil {
			return float64Pointer(agg)
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("aggregation error version %v, metric %v, and aggregation func %v", i, baseMetric, a)
			return nil
		}
	case MinAggregator:
		agg, err := stats.Min(vals)
		if err == nil {
			return float64Pointer(agg)
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("aggregation error version %v, metric %v, and aggregation func %v", i, baseMetric, a)
			return nil
		}
	case MaxAggregator:
		agg, err := stats.Mean(vals)
		if err == nil {
			return float64Pointer(agg)
		} else {
			log.Logger.WithStackTrace(err.Error()).Error("aggregation error version %v, metric %v, and aggregation func %v", i, baseMetric, a)
			return nil
		}
	default: // don't do anything
	}

	// at this point, 'a' must be a percentile aggregator
	if strings.HasPrefix(a, "p") {
		b := strings.TrimPrefix(a, "p")
		// b must be a percent
		if match, _ := regexp.MatchString(decimalRegex, b); match {
			// extract percent
			if percent, err := strconv.ParseFloat(b, 64); err != nil {
				log.Logger.WithStackTrace(err.Error()).Error("error extracting percent from aggregation func %v", a)
				return nil
			} else {
				// compute percentile
				agg, err := stats.Percentile(vals, percent)
				if err == nil {
					return float64Pointer(agg)
				} else {
					log.Logger.WithStackTrace(err.Error()).Error("aggregation error version %v, metric %v, and aggregation func %v", i, baseMetric, a)
					return nil
				}
			}
		} else {
			log.Logger.Errorf("unable to extract percent from agggregation func %v", a)
			return nil
		}
	} else {
		log.Logger.Errorf("invalid aggregation %v", a)
		return nil
	}
}

// aggregateMetric returns the aggregated metric value for a given version and metric
func (in *Insights) aggregateMetric(i int, m string) *float64 {
	s := strings.Split(m, "/")
	baseMetric := s[0] + "/" + s[1]
	if m, ok := in.MetricsInfo[baseMetric]; ok {
		log.Logger.Tracef("found metric %v used for aggregation", baseMetric)
		if m.Type == SampleMetricType {
			log.Logger.Tracef("metric %v used for aggregation is a sample metric", baseMetric)
			return in.getSampleAggregation(i, baseMetric, s[2])
		} else {
			log.Logger.Errorf("metric %v used for aggregation is not a sample metric", baseMetric)
			return nil
		}
	} else {
		log.Logger.Warnf("could not find metric %v used for aggregation", baseMetric)
		return nil
	}
}

// normalizeMetricName normalizes percentile values in metric names
func normalizeMetricName(m string) (string, error) {
	pre := iter8BuiltInPrefix + "/" + builtInHTTPLatencyPercentilePrefix
	if strings.HasPrefix(m, pre) { // built-in http percentile metric
		remainder := strings.TrimPrefix(m, pre)
		if percent, e := strconv.ParseFloat(remainder, 64); e != nil {
			err := fmt.Errorf("cannot extract percent from metric %v", m)
			log.Logger.WithStackTrace(e.Error()).Error(err)
			return m, err
		} else {
			// return percent normalized metric name
			return fmt.Sprintf("%v%v", pre, percent), nil
		}
	} else {
		// not a built-in http percentile metric
		return m, nil
	}
}

// getMetricValue gets the value of the given metric for the given version
func (in *Insights) getMetricValue(i int, m string) *float64 {
	s := strings.Split(m, "/")
	if len(s) == 3 {
		log.Logger.Tracef("%v is an aggregated metric", m)
		return in.aggregateMetric(i, m)
	} else if len(s) == 2 { // this appears to be a non-aggregated metric
		if nm, err := normalizeMetricName(m); err != nil {
			return nil
		} else {
			return in.getMetricFromValuesMap(i, nm)
		}
	} else {
		log.Logger.Errorf("invalid metric name %v", m)
		log.Logger.Error("metric names must be of the form a/b or a/b/c, where a is the id of the metrics backend, b is the id of a metric name, and c is a valid aggregation function")
		return nil
	}
}
