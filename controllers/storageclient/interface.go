// Package storageclient provides the storage client for the controllers package
package storageclient

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/montanaflynn/stats"
	"gonum.org/v1/plot/plotter"
)

// SummarizedMetric is a metric summary
type SummarizedMetric struct {
	Count  uint64
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

// VersionMetrics contains all the metrics over transactions and over users
// key = metric name
type VersionMetrics map[string]struct {
	MetricsOverTransactions []float64
	MetricsOverUsers        []float64
}

// Interface enables interaction with a storage entity
// Can be mocked in unit tests with fake implementation
type Interface interface {
	// Returns a nested map of the metrics data for a particular application, version, and signature
	// Example:
	//	{
	//		"my-metric": {
	//			"MetricsOverTransactions": [1, 1, 3, 4, 5]
	//			"MetricsOverUsers": [2, 7, 5]
	//		}
	//	}
	//
	// NOTE: for users that have not produced any metrics (for example, via lookup()), GetMetrics() will add 0s for the extra users in metricsOverUsers
	// Example, given 5 total users:
	//
	//	{
	//		"my-metric": {
	//			"MetricsOverTransactions": [1, 1, 3, 4, 5]
	//			"MetricsOverUsers": [2, 7, 5, 0, 0]
	//		}
	//	}
	GetMetrics(applicationName string, version int, signature string) (*VersionMetrics, error)

	// called by the A/B/n SDK gRPC API implementation (SDK for application clients)
	// Example key: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> my-metric-value (get the metric value with all the provided information)
	SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error

	// Example key: kt-users::my-app::0::my-signature::my-user -> true
	SetUser(applicationName string, version int, signature, user string) error
}

//
//
// Test Grafana
//
//

type testMetricVersionSummaryConfig struct {
	numPoints int
	mean      float64
	stdDev    float64
}

type metricConfig struct {
	numBuckets     int
	decimalPlace   float64
	versionConfigs []testMetricVersionSummaryConfig
}

type grafanaConfig struct {
	seed          int64
	metricConfigs map[string]metricConfig
}

// VersionSummarizedMetric adds version to summary data
type VersionSummarizedMetric struct {
	Version int
	SummarizedMetric
}

// calculateSummarizedMetric calculates a metric summary for a particular collection of data
func calculateSummarizedMetric(data []float64) (SummarizedMetric, error) {
	if len(data) == 0 {
		return SummarizedMetric{}, nil
	}

	// NOTE: len() does not produce a uint64
	count := uint64(len(data))

	min, err := stats.Min(data)
	if err != nil {
		return SummarizedMetric{}, err
	}

	max, err := stats.Max(data)
	if err != nil {
		return SummarizedMetric{}, err
	}

	mean, err := stats.Mean(data)
	if err != nil {
		return SummarizedMetric{}, err
	}

	stdDev, err := stats.StandardDeviation(data)
	if err != nil {
		return SummarizedMetric{}, err
	}

	return SummarizedMetric{
		Count:  count,
		Mean:   mean,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
	}, nil
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

// VersionMetricSummary is a metric summary for a given app version
type VersionMetricSummary struct {
	NumUsers uint64

	// key = metric name; value is the metric summary
	MetricSummaries map[string]MetricSummary
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

// getTestGrafanaHelper returns the Iter8 Grafana dashboard for a single metric
func getTestGrafanaHelper(c metricConfig, r *rand.Rand) (*MetricSummary, error) {
	transactions := map[string][]float64{}
	users := map[string][]float64{}

	summaryOverTransactions := []*VersionSummarizedMetric{}
	summaryOverUsers := []*VersionSummarizedMetric{}

	for i, versionConfig := range c.versionConfigs {
		version := fmt.Sprintf("%d", i)

		// create transaction data set
		transactions[version] = []float64{}
		for j := 0; j < versionConfig.numPoints; j++ {
			normFloat := rand.NormFloat64()
			if r != nil {
				normFloat = r.NormFloat64()
			}

			transactions[version] = append(transactions[version], versionConfig.mean+(normFloat*versionConfig.stdDev))
		}

		// summary for transactions
		summarizedMetric, err := calculateSummarizedMetric(transactions[version])
		if err != nil {
			return nil, err
		}
		summaryOverTransactions = append(summaryOverTransactions, &VersionSummarizedMetric{
			Version:          i,
			SummarizedMetric: summarizedMetric,
		})

		// create users data set
		// combine every other data point
		users[version] = []float64{}
		for k := 0; k < len(transactions[version])/2; k++ {
			users[version] = append(users[version], transactions[version][k*2]+transactions[version][(k*2)+1])

			if (k == (len(transactions[version])/2)-1) && (len(transactions[version])%2 == 1) {
				users[version] = append(users[version], transactions[version][(k*2)+2])

				break
			}
		}

		// summary for users
		summarizedMetricUser, err := calculateSummarizedMetric(users[version])
		if err != nil {
			return nil, err
		}
		summaryOverUsers = append(summaryOverUsers, &VersionSummarizedMetric{
			Version:          i,
			SummarizedMetric: summarizedMetricUser,
		})
	}

	histogramsOverTransactions, err := calculateHistogram(transactions, c.numBuckets, c.decimalPlace)
	if err != nil {
		return nil, err
	}

	histogramsOverUsers, err := calculateHistogram(users, c.numBuckets, c.decimalPlace)
	if err != nil {
		return nil, err
	}

	return &MetricSummary{
		HistogramsOverTransactions: &histogramsOverTransactions,
		HistogramsOverUsers:        &histogramsOverUsers,
		SummaryOverTransactions:    summaryOverTransactions,
		SummaryOverUsers:           summaryOverUsers,
	}, nil
}

// getTestGrafana returns a test for the Iter8 Grafana dashboard
func getTestGrafana(c grafanaConfig) (map[string]*MetricSummary, error) {
	res := map[string]*MetricSummary{}

	r := &rand.Rand{}
	if c.seed != 0 {
		r = rand.New(rand.NewSource(c.seed))
	}

	for metric, metricConfig := range c.metricConfigs {
		x, err := getTestGrafanaHelper(metricConfig, r)
		if err != nil {
			return nil, err
		}

		res[metric] = x
	}

	return res, nil
}
