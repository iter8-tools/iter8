package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/iter8-tools/iter8/abn"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/storage"
	"github.com/montanaflynn/stats"
	"gonum.org/v1/plot/plotter"

	"fortio.org/fortio/fhttp"
	fstats "fortio.org/fortio/stats"
)

const (
	configEnv         = "METRICS_CONFIG_FILE"
	defaultPortNumber = 8080
	timeFormat        = "02 Jan 06 15:04 MST"
)

// metricsConfig defines the configuration of the controllers
type metricsConfig struct {
	// Port is port number on which the metrics service should listen
	Port *int `json:"port,omitempty"`
}

// versionSummarizedMetric adds version to summary data
type versionSummarizedMetric struct {
	Version int
	storage.SummarizedMetric
}

// grafanaHistogram represents the histogram in the Grafana Iter8 dashboard
type grafanaHistogram []grafanaHistogramBucket

// grafanaHistogramBucket represents a bucket in the histogram in the Grafana Iter8 dashboard
type grafanaHistogramBucket struct {
	// Version is the version of the application
	Version string

	// Bucket is the bucket of the histogram
	// For example: 8-12
	Bucket string

	// Value is the number of points in this bucket
	Value float64
}

// metricSummary is result for a metric
type metricSummary struct {
	HistogramsOverTransactions *grafanaHistogram
	HistogramsOverUsers        *grafanaHistogram
	SummaryOverTransactions    []*versionSummarizedMetric
	SummaryOverUsers           []*versionSummarizedMetric
}

// dashboardExperimentResult is a capitalized version of ExperimentResult used to display data in Grafana
type dashboardExperimentResult struct {
	// Name is the name of this experiment
	Name string

	// Namespace is the namespace of this experiment
	Namespace string

	// Revision of this experiment
	Revision int

	// StartTime is the time when the experiment run started
	StartTime string `json:"Start time"`

	// NumCompletedTasks is the number of completed tasks
	NumCompletedTasks int `json:"Completed tasks"`

	// Failure is true if any of its tasks failed
	Failure bool

	// Insights produced in this experiment
	Insights *util.Insights

	// Iter8Version is the version of Iter8 CLI that created this result object
	Iter8Version string `json:"Iter8 version"`
}

// httpEndpointRow is the data needed to produce a single row for an HTTP experiment in the Iter8 Grafana dashboard
type httpEndpointRow struct {
	Durations  grafanaHistogram
	Statistics storage.SummarizedMetric

	ErrorDurations  grafanaHistogram         `json:"Error durations"`
	ErrorStatistics storage.SummarizedMetric `json:"Error statistics"`

	ReturnCodes map[int]int64 `json:"Return codes"`
}

type httpDashboard struct {
	// key is the endpoint
	Endpoints map[string]httpEndpointRow

	ExperimentResult dashboardExperimentResult
}

type ghzStatistics struct {
	Count      uint64
	ErrorCount float64
}

// ghzEndpointRow is the data needed to produce a single row for an gRPC experiment in the Iter8 Grafana dashboard
type ghzEndpointRow struct {
	Durations              grafanaHistogram
	Statistics             ghzStatistics
	StatusCodeDistribution map[string]int `json:"Status codes"`
}

type ghzDashboard struct {
	// key is the endpoint
	Endpoints map[string]ghzEndpointRow

	ExperimentResult dashboardExperimentResult
}

var allRoutemaps controllers.AllRouteMapsInterface = &controllers.DefaultRoutemaps{}

// Start starts the HTTP server
func Start(stopCh <-chan struct{}) error {
	// read configutation for metrics service
	conf := &metricsConfig{}
	err := util.ReadConfig(configEnv, conf, func() {
		if nil == conf.Port {
			conf.Port = util.IntPointer(defaultPortNumber)
		}
	})
	if err != nil {
		log.Logger.Errorf("unable to read metrics configuration: %s", err.Error())
		return err
	}

	// configure endpoints
	http.HandleFunc(util.TestResultPath, putExperimentResult)
	http.HandleFunc(util.AbnDashboard, getAbnDashboard)
	http.HandleFunc(util.HTTPDashboardPath, getHTTPDashboard)
	http.HandleFunc(util.GRPCDashboardPath, getGRPCDashboard)

	// configure HTTP server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", *conf.Port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		<-stopCh
		log.Logger.Warnf("stop channel closed, shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// start HTTP server
	err = server.ListenAndServe()
	if err != nil {
		log.Logger.Errorf("unable to start metrics service: %s", err.Error())
		return err
	}

	return nil
}

// getAbnDashboard handles GET /abnDashboard with query parameter application=name and namespace=namespace
func getAbnDashboard(w http.ResponseWriter, r *http.Request) {
	log.Logger.Trace("getAbnDashboard called")
	defer log.Logger.Trace("getAbnDashboard completed")

	// verify method
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameters)
	application := r.URL.Query().Get("application")
	if application == "" {
		http.Error(w, "no application specified", http.StatusBadRequest)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "no namespace specified", http.StatusBadRequest)
		return
	}

	namespaceApplication := fmt.Sprintf("%s/%s", namespace, application)

	log.Logger.Tracef("getAbnDashboard called for application %s", namespaceApplication)

	// identify the routemap for the application
	rm := allRoutemaps.GetAllRoutemaps().GetRoutemapFromNamespaceName(namespace, application)
	if rm == nil || reflect.ValueOf(rm).IsNil() {
		http.Error(w, fmt.Sprintf("unknown application %s", namespaceApplication), http.StatusBadRequest)
		return
	}
	log.Logger.Tracef("getAbnDashboard found routemap %v", rm)

	// initialize result
	result := make(map[string]*metricSummary, 0)
	byMetricOverTransactions := make(map[string](map[string][]float64), 0)
	byMetricOverUsers := make(map[string](map[string][]float64), 0)

	// for each version:
	//   get metrics
	//   for each metric, compute summary for metric, version
	//   prepare for histogram computation
	for v, version := range rm.GetVersions() {
		signature := version.GetSignature()
		if signature == nil {
			log.Logger.Debugf("no signature for application %s (version %d)", namespaceApplication, v)
			continue
		}

		if abn.MetricsClient == nil {
			log.Logger.Error("no metrics client")
			continue
		}
		versionmetrics, err := abn.MetricsClient.GetMetrics(namespaceApplication, v, *signature)
		if err != nil {
			log.Logger.Debugf("no metrics found for application %s (version %d; signature %s)", namespaceApplication, v, *signature)
			continue
		}

		for metric, metrics := range *versionmetrics {
			_, ok := result[metric]
			if !ok {
				// no entry for metric result; create empty entry
				result[metric] = &metricSummary{
					HistogramsOverTransactions: nil,
					HistogramsOverUsers:        nil,
					SummaryOverTransactions:    []*versionSummarizedMetric{},
					SummaryOverUsers:           []*versionSummarizedMetric{},
				}
			}

			entry := result[metric]

			smT, err := calculateSummarizedMetric(metrics.MetricsOverTransactions)
			if err != nil {
				log.Logger.Debugf("unable to compute summaried metrics over transactions for application %s (version %d; signature %s)", namespaceApplication, v, *signature)
				continue
			} else {
				entry.SummaryOverTransactions = append(entry.SummaryOverTransactions, &versionSummarizedMetric{
					Version:          v,
					SummarizedMetric: smT,
				})
			}

			smU, err := calculateSummarizedMetric(metrics.MetricsOverUsers)
			if err != nil {
				log.Logger.Debugf("unable to compute summaried metrics over users for application %s (version %d; signature %s)", namespaceApplication, v, *signature)
				continue
			}
			entry.SummaryOverUsers = append(entry.SummaryOverUsers, &versionSummarizedMetric{
				Version:          v,
				SummarizedMetric: smU,
			})
			result[metric] = entry

			// copy data into structure for histogram calculation (to be done later)
			vStr := fmt.Sprintf("%d", v)
			// over transaction data
			_, ok = byMetricOverTransactions[metric]
			if !ok {
				byMetricOverTransactions[metric] = make(map[string][]float64, 0)
			}
			(byMetricOverTransactions[metric])[vStr] = metrics.MetricsOverTransactions

			// over user data
			_, ok = byMetricOverUsers[metric]
			if !ok {
				byMetricOverUsers[metric] = make(map[string][]float64, 0)
			}
			(byMetricOverUsers[metric])[vStr] = metrics.MetricsOverUsers
		}
	}

	// compute histograms
	for metric, byVersion := range byMetricOverTransactions {
		hT, err := calculateHistogram(byVersion, 0, 0)
		if err != nil {
			log.Logger.Debugf("unable to compute histogram over transactions for application %s (metric %s)", namespaceApplication, metric)
			continue
		} else {
			resultEntry := result[metric]
			resultEntry.HistogramsOverTransactions = &hT
			result[metric] = resultEntry
		}
	}

	for metric, byVersion := range byMetricOverUsers {
		hT, err := calculateHistogram(byVersion, 0, 0)
		if err != nil {
			log.Logger.Debugf("unable to compute histogram over users for application %s (metric %s)", namespaceApplication, metric)
			continue
		} else {
			resultEntry := result[metric]
			resultEntry.HistogramsOverUsers = &hT
			result[metric] = resultEntry
		}
	}

	// convert to JSON
	b, err := json.MarshalIndent(result, "", "   ")
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to create JSON response %s", string(b)), http.StatusInternalServerError)
		return
	}

	// finally, send response
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// calculateSummarizedMetric calculates a metric summary for a particular collection of data
func calculateSummarizedMetric(data []float64) (storage.SummarizedMetric, error) {
	if len(data) == 0 {
		return storage.SummarizedMetric{}, nil
	}

	// NOTE: len() does not produce a uint64
	count := uint64(len(data))

	min, err := stats.Min(data)
	if err != nil {
		return storage.SummarizedMetric{}, err
	}

	max, err := stats.Max(data)
	if err != nil {
		return storage.SummarizedMetric{}, err
	}

	mean, err := stats.Mean(data)
	if err != nil {
		return storage.SummarizedMetric{}, err
	}

	stdDev, err := stats.StandardDeviation(data)
	if err != nil {
		return storage.SummarizedMetric{}, err
	}

	return storage.SummarizedMetric{
		Count:  count,
		Mean:   mean,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
	}, nil
}

// calculateHistogram creates histograms for multiple versions
// the histograms have the same buckets so they can be displayed together
// numBuckets is the number of buckets in the histogram
// decimalPlace is the number of decimal places that the histogram labels should be rounded to
//
//	For example: "-0.24178488465151116 - 0.24782423875427073" -> "-0.242 - 0.248"
//
// TODO: defaults for numBuckets/decimalPlace?
func calculateHistogram(versionMetrics map[string][]float64, numBuckets int, decimalPlace float64) (grafanaHistogram, error) {
	if numBuckets == 0 {
		numBuckets = 10
	}
	if decimalPlace == 0 {
		decimalPlace = 1
	}

	mins := []float64{}
	maxs := []float64{}
	for _, metrics := range versionMetrics {
		summary, err := calculateSummarizedMetric(metrics)
		if err != nil {
			return nil, fmt.Errorf("cannot calculate summarized metric: %e", err)
		}

		mins = append(mins, summary.Min)
		maxs = append(maxs, summary.Max)
	}

	// versionMin is the minimum across all versions
	// versionMax is the maximum across all versions
	// added to the metrics of each version in order to ensure consistent bins across all versions
	versionMin, err := stats.Min(mins)
	if err != nil {
		return nil, fmt.Errorf("cannot calculate version minimum: %e", err)
	}
	versionMax, err := stats.Max(maxs)
	if err != nil {
		return nil, fmt.Errorf("cannot create version maximum: %e", err)
	}

	grafanaHistogram := grafanaHistogram{}

	for version, metrics := range versionMetrics {
		// convert the raw values to the gonum plot values
		values := make(plotter.Values, len(metrics))
		copy(values, metrics)

		// append the minimum and maximum across all versions
		// allows all the buckets to be the same across versions
		values = append(values, versionMin, versionMax)

		h, err := plotter.NewHist(values, numBuckets)
		if err != nil {
			return nil, fmt.Errorf("cannot create Grafana historgram: %e", err)
		}

		for i, bin := range h.Bins {
			count := bin.Weight
			// reduce the count for the starting and ending bins to compensate for versionMin and versionMax
			// bins are sorted by bucket
			// TODO: verify bins are sorted
			if i == 0 || i == len(h.Bins)-1 {
				count--
			}

			grafanaHistogram = append(grafanaHistogram, grafanaHistogramBucket{
				Version: version,
				Bucket:  bucketLabel(bin.Min, bin.Max, decimalPlace),
				Value:   count,
			})
		}
	}

	return grafanaHistogram, nil
}

// roundDecimal rounds a given number to the given decimal place
// For example: if x = 2270424855658346, decimalPlace = 3, then return 1.227
func roundDecimal(x float64, decimalPlace float64) float64 {
	y := math.Pow(10, decimalPlace)

	return math.Floor(x*y) / y
}

// bucketLabel return a label for a histogram bucket
func bucketLabel(min, max float64, decimalPlace float64) string {
	return fmt.Sprintf("%s - %s", strconv.FormatFloat(roundDecimal(min, decimalPlace), 'f', -1, 64), strconv.FormatFloat(roundDecimal(max, decimalPlace), 'f', -1, 64))
}

func getHTTPHistogram(fortioHistogram []fstats.Bucket, decimalPlace float64) grafanaHistogram {
	grafanaHistogram := grafanaHistogram{}

	for _, bucket := range fortioHistogram {
		grafanaHistogram = append(grafanaHistogram, grafanaHistogramBucket{
			Version: "0",
			Bucket:  bucketLabel(bucket.Start*1000, bucket.End*1000, decimalPlace),
			Value:   float64(bucket.Count),
		})
	}

	return grafanaHistogram
}

func getHTTPStatistics(fortioHistogram *fstats.HistogramData, decimalPlace float64) storage.SummarizedMetric {
	return storage.SummarizedMetric{
		Count:  uint64(fortioHistogram.Count),
		Mean:   fortioHistogram.Avg * 1000,
		StdDev: fortioHistogram.StdDev * 1000,
		Min:    fortioHistogram.Min * 1000,
		Max:    fortioHistogram.Max * 1000,
	}
}

func getHTTPEndpointRow(httpRunnerResults *fhttp.HTTPRunnerResults) httpEndpointRow {
	row := httpEndpointRow{}
	if httpRunnerResults.DurationHistogram != nil {
		row.Durations = getHTTPHistogram(httpRunnerResults.DurationHistogram.Data, 1)
		row.Statistics = getHTTPStatistics(httpRunnerResults.DurationHistogram, 1)
	}

	if httpRunnerResults.ErrorsDurationHistogram != nil {
		row.ErrorDurations = getHTTPHistogram(httpRunnerResults.ErrorsDurationHistogram.Data, 1)
		row.ErrorStatistics = getHTTPStatistics(httpRunnerResults.ErrorsDurationHistogram, 1)
	}

	row.ReturnCodes = httpRunnerResults.RetCodes

	return row
}

func getHTTPDashboardHelper(experimentResult *util.ExperimentResult) httpDashboard {
	dashboard := httpDashboard{
		Endpoints: map[string]httpEndpointRow{},
		ExperimentResult: dashboardExperimentResult{
			Name:              experimentResult.Name,
			Namespace:         experimentResult.Namespace,
			Revision:          experimentResult.Revision,
			StartTime:         experimentResult.StartTime.Time.Format(timeFormat),
			NumCompletedTasks: experimentResult.NumCompletedTasks,
			Failure:           experimentResult.Failure,
			Iter8Version:      experimentResult.Iter8Version,
		},
	}

	// get raw data from ExperimentResult
	httpTaskData := experimentResult.Insights.TaskData[util.CollectHTTPTaskName]
	if httpTaskData == nil {
		log.Logger.Error("cannot get http task data from Insights")
		return dashboard
	}

	httpTaskDataBytes, err := json.Marshal(httpTaskData)
	if err != nil {
		log.Logger.Error("cannot marshal http task data")
		return dashboard
	}

	httpResult := util.HTTPResult{}
	err = json.Unmarshal(httpTaskDataBytes, &httpResult)
	if err != nil {
		log.Logger.Error("cannot unmarshal http task data into HTTPResult")
		return dashboard
	}

	// form rows of dashboard
	for endpoint, endpointResult := range httpResult {
		endpointResult := endpointResult
		dashboard.Endpoints[endpoint] = getHTTPEndpointRow(endpointResult)
	}

	return dashboard
}

// getHTTPDashboard handles GET /getHTTPDashboard with query parameter test=name and namespace=namespace
func getHTTPDashboard(w http.ResponseWriter, r *http.Request) {
	log.Logger.Trace("getHTTPGrafana called")
	defer log.Logger.Trace("getHTTPGrafana completed")

	// verify method
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameters)
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "no namespace specified", http.StatusBadRequest)
		return
	}

	test := r.URL.Query().Get("test")
	if test == "" {
		http.Error(w, "no test specified", http.StatusBadRequest)
		return
	}

	log.Logger.Tracef("getHTTPGrafana called for namespace %s and test %s", namespace, test)

	// get fortioResult from metrics client
	if abn.MetricsClient == nil {
		http.Error(w, "no metrics client", http.StatusInternalServerError)
		return
	}

	// get testResult from metrics client
	testResult, err := abn.MetricsClient.GetExperimentResult(namespace, test)
	if err != nil {
		errorMessage := fmt.Sprintf("cannot get experiment result with namespace %s, test %s", namespace, test)
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// JSON marshal the dashboard
	dashboardBytes, err := json.Marshal(getHTTPDashboardHelper(testResult))
	if err != nil {
		errorMessage := "cannot JSON marshal HTTP dashboard"
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	// finally, send response
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(dashboardBytes)
}

func getGRPCHistogram(ghzHistogram []runner.Bucket, decimalPlace float64) grafanaHistogram {
	grafanaHistogram := grafanaHistogram{}

	for _, bucket := range ghzHistogram {
		grafanaHistogram = append(grafanaHistogram, grafanaHistogramBucket{
			Version: "0",
			Bucket:  fmt.Sprint(roundDecimal(bucket.Mark*1000, 3)),
			Value:   float64(bucket.Count),
		})
	}

	return grafanaHistogram
}

func getGRPCStatistics(ghzRunnerReport *runner.Report) ghzStatistics {
	// populate error count & rate
	ec := float64(0)
	for _, count := range ghzRunnerReport.ErrorDist {
		ec += float64(count)
	}

	return ghzStatistics{
		Count:      ghzRunnerReport.Count,
		ErrorCount: ec,
	}
}

func getGRPCEndpointRow(ghzRunnerReport *runner.Report) ghzEndpointRow {
	row := ghzEndpointRow{}

	if ghzRunnerReport.Histogram != nil {
		row.Durations = getGRPCHistogram(ghzRunnerReport.Histogram, 3)
		row.Statistics = getGRPCStatistics(ghzRunnerReport)
	}

	row.StatusCodeDistribution = ghzRunnerReport.StatusCodeDist

	return row
}

func getGRPCDashboardHelper(experimentResult *util.ExperimentResult) ghzDashboard {
	dashboard := ghzDashboard{
		Endpoints: map[string]ghzEndpointRow{},
		ExperimentResult: dashboardExperimentResult{
			Name:              experimentResult.Name,
			Namespace:         experimentResult.Namespace,
			Revision:          experimentResult.Revision,
			StartTime:         experimentResult.StartTime.Time.Format(timeFormat),
			NumCompletedTasks: experimentResult.NumCompletedTasks,
			Failure:           experimentResult.Failure,
			Iter8Version:      experimentResult.Iter8Version,
		},
	}

	// get raw data from ExperimentResult
	ghzTaskData := experimentResult.Insights.TaskData[util.CollectGRPCTaskName]
	if ghzTaskData == nil {
		return dashboard
	}

	ghzTaskDataBytes, err := json.Marshal(ghzTaskData)
	if err != nil {
		log.Logger.Error("cannot marshal ghz task data")
		return dashboard
	}

	ghzResult := util.GHZResult{}
	err = json.Unmarshal(ghzTaskDataBytes, &ghzResult)
	if err != nil {
		log.Logger.Error("cannot unmarshal ghz task data into GHZResult")
		return dashboard
	}

	// form rows of dashboard
	for endpoint, endpointResult := range ghzResult {
		endpointResult := endpointResult
		dashboard.Endpoints[endpoint] = getGRPCEndpointRow(endpointResult)
	}

	return dashboard
}

// getGRPCDashboard handles GET /getGRPCDashboard with query parameter test=name and namespace=namespace
func getGRPCDashboard(w http.ResponseWriter, r *http.Request) {
	log.Logger.Trace("getGRPCDashboard called")
	defer log.Logger.Trace("getGRPCDashboard completed")

	// verify method
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameters)
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "no namespace specified", http.StatusBadRequest)
		return
	}

	test := r.URL.Query().Get("test")
	if test == "" {
		http.Error(w, "no test specified", http.StatusBadRequest)
		return
	}

	log.Logger.Tracef("getGRPCDashboard called for namespace %s and test %s", namespace, test)

	// get ghz result from metrics client
	if abn.MetricsClient == nil {
		http.Error(w, "no metrics client", http.StatusInternalServerError)
		return
	}

	// get testResult from metrics client
	testResult, err := abn.MetricsClient.GetExperimentResult(namespace, test)
	if err != nil {
		errorMessage := fmt.Sprintf("cannot get experiment result with namespace %s, test %s", namespace, test)
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// JSON marshal the dashboard
	dashboardBytes, err := json.Marshal(getGRPCDashboardHelper(testResult))
	if err != nil {
		errorMessage := "cannot JSON marshal gRPC dashboard"
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	// finally, send response
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(dashboardBytes)
}

// putExperimentResult handles PUT /testResult with query parameter test=name and namespace=namespace
func putExperimentResult(w http.ResponseWriter, r *http.Request) {
	log.Logger.Trace("putExperimentResult called")
	defer log.Logger.Trace("putExperimentResult completed")

	// verify method
	if r.Method != http.MethodPut {
		http.Error(w, "expected PUT", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameters)
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "no namespace specified", http.StatusBadRequest)
		return
	}

	experiment := r.URL.Query().Get("test")
	if experiment == "" {
		http.Error(w, "no experiment specified", http.StatusBadRequest)
		return
	}

	log.Logger.Tracef("putExperimentResult called for namespace %s and test %s", namespace, experiment)

	defer func() {
		err := r.Body.Close()
		if err != nil {
			errorMessage := fmt.Sprintf("cannot close request body: %e", err)
			log.Logger.Error(errorMessage)
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}
	}()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorMessage := fmt.Sprintf("cannot read request body: %e", err)
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	experimentResult := util.ExperimentResult{}
	err = json.Unmarshal(body, &experimentResult)
	if err != nil {
		errorMessage := fmt.Sprintf("cannot unmarshal body into ExperimentResult: %s: %e", string(body), err)
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	if abn.MetricsClient == nil {
		http.Error(w, "no metrics client", http.StatusInternalServerError)
		return
	}

	err = abn.MetricsClient.SetExperimentResult(namespace, experiment, &experimentResult)
	if err != nil {
		errorMessage := fmt.Sprintf("cannot store result in storage client: %s: %e", string(body), err)
		log.Logger.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	// TODO: 201 for new resource, 200 for update
}
