package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/iter8-tools/iter8/abn"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/storage"
	"github.com/montanaflynn/stats"
	"gonum.org/v1/plot/plotter"
)

type configMaps interface {
	getAllConfigMaps() controllers.RoutemapsInterface
}

type defaultConfigMaps struct{}

func (cm *defaultConfigMaps) getAllConfigMaps() controllers.RoutemapsInterface {
	return &controllers.AllRoutemaps
}

var allConfigMaps configMaps = &defaultConfigMaps{}

// Start starts the HTTP server
func Start() error {
	http.HandleFunc("/metrics", getMetrics)
	// http.HandleFunc("/summarymetrics", getSummaryMetrics)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Logger.Errorf("unable to start metrics service: %s", err.Error())
		return err
	}
	return nil
}

// VersionSummarizedMetric adds version to summary data
type VersionSummarizedMetric struct {
	Version int
	storage.SummarizedMetric
}

// GrafanaHistogram represents the histogram in the Grafana Iter8 dashboard
type GrafanaHistogram []GrafanaHistogramBucket

// GrafanaHistogramBucket represents a bucket in the histogram in the Grafana Iter8 dashboard
type GrafanaHistogramBucket struct {
	// Version is the version of the application
	Version string

	// Bucket is the bucket of the histogram
	// For example: 8-12
	Bucket string

	// Value is the number of points in this bucket
	Value float64
}

// MetricSummary is result for a metric
type MetricSummary struct {
	HistogramsOverTransactions *GrafanaHistogram
	HistogramsOverUsers        *GrafanaHistogram
	SummaryOverTransactions    []*VersionSummarizedMetric
	SummaryOverUsers           []*VersionSummarizedMetric
}

// getMetrics handles POST /metrics
func getMetrics(w http.ResponseWriter, r *http.Request) {
	log.Logger.Trace("getMetrics called")
	defer log.Logger.Trace("getMetrics completed")

	// verify method
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameter)
	application := r.URL.Query().Get("application")
	if application == "" {
		http.Error(w, "no application specified", http.StatusBadRequest)
	}
	log.Logger.Tracef("getMetrics called for application %s", application)

	// identify the routemap for the application
	namespace, name := splitApplicationKey(application)
	rm := allConfigMaps.getAllConfigMaps().GetRoutemapFromNamespaceName(namespace, name)
	// rm := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(namespace, name)
	if rm == nil {
		http.Error(w, fmt.Sprintf("unknown application %s", application), http.StatusBadRequest)
		return
	}
	log.Logger.Trace("getMetrics found routemap ", rm)

	// initialize result
	result := make(map[string]MetricSummary, 0)
	byMetricOverTransactions := make(map[string](map[string][]float64), 0)
	byMetricOverUsers := make(map[string](map[string][]float64), 0)

	// for each version:
	//   get metrics
	//   for each metric, compute summary for metric, version
	//   prepare for histogram computation
	numVersions := len(rm.GetVersions())
	for v, version := range rm.GetVersions() {
		signature := version.GetSignature()
		if signature == nil {
			log.Logger.Debugf("no signature for application %s (version %d)", application, v)
			continue
		}

		versionmetrics, err := abn.MetricsClient.GetMetrics(application, v, *signature)
		if err != nil {
			log.Logger.Debugf("no metrics found for application %s (version %d; signature %s)", application, v, *signature)
			continue
		}

		for metric, metrics := range *versionmetrics {
			_, ok := result[metric]
			if !ok {
				// no entry for metric result; create empty entry
				result[metric] = MetricSummary{
					HistogramsOverTransactions: nil,
					HistogramsOverUsers:        nil,
					SummaryOverTransactions:    make([]*VersionSummarizedMetric, numVersions),
					SummaryOverUsers:           make([]*VersionSummarizedMetric, numVersions),
				}
			}

			smT, err := calculateSummarizedMetric(metrics.MetricsOverTransactions)
			if err != nil {
				log.Logger.Debugf("unable to compute summaried metrics over transactions for application %s (version %d; signature %s)", application, v, *signature)
				continue
			} else {
				result[metric].SummaryOverTransactions[v] = &VersionSummarizedMetric{
					Version:          v,
					SummarizedMetric: smT,
				}
			}

			smU, err := calculateSummarizedMetric(metrics.MetricsOverUsers)
			if err != nil {
				log.Logger.Debugf("unable to compute summaried metrics over users for application %s (version %d; signature %s)", application, v, *signature)
				continue
			}
			result[metric].SummaryOverUsers[v] = &VersionSummarizedMetric{
				Version:          v,
				SummarizedMetric: smU,
			}

			// copy data into structure for histogram calculation (to be done later)
			// over transaction data
			vStr := fmt.Sprintf("%d", v)
			_, ok = byMetricOverTransactions[metric]
			if !ok {
				byMetricOverTransactions[metric] = make(map[string][]float64, 0)
			}
			(byMetricOverTransactions[metric])[vStr] = metrics.MetricsOverTransactions

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
			log.Logger.Debugf("unable to compute histogram over transactions for application %s (metric %s)", application, metric)
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
			log.Logger.Debugf("unable to compute histogram over users for application %s (metric %s)", application, metric)
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

// splitApplicationKey is a utility function that returns the name and namespace from a key of the form "namespace/name"
func splitApplicationKey(applicationKey string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
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
func calculateHistogram(versionMetrics map[string][]float64, numBuckets int, decimalPlace float64) (GrafanaHistogram, error) {
	if numBuckets == 0 {
		numBuckets = 20
	}
	if decimalPlace == 0 {
		decimalPlace = 3
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

	grafanaHistogram := GrafanaHistogram{}

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

			grafanaHistogram = append(grafanaHistogram, GrafanaHistogramBucket{
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
