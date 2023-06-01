package report

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sort"

	htmlT "html/template"

	_ "embed"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

// HTMLReporter supports generation of HTML reports from experiments.
type HTMLReporter struct {
	// Reporter enables access to all reporter data and methods
	*Reporter
}

// reportHTML is the HTML report template
//
//go:embed htmlreport.tpl
var reportHTML string

// Gen creates an HTML report for a given experiment
func (ht *HTMLReporter) Gen(out io.Writer) error {

	// create HTML template
	htpl, err := htmlT.New("report").Option("missingkey=error").Funcs(sprig.FuncMap()).Funcs(htmlT.FuncMap{
		"renderSLOSatisfiedHTML":      renderSLOSatisfiedHTML,
		"renderSLOSatisfiedCellClass": renderSLOSatisfiedCellClass,
	}).Parse(reportHTML)
	if err != nil {
		e := errors.New("unable to parse HTML template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	var b bytes.Buffer
	if err = htpl.Execute(&b, ht); err != nil {
		e := errors.New("unable to execute template")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		return e
	}

	// print output
	fmt.Fprintln(out, b.String())
	return nil
}

// RenderStr is a helper method for rendering strings
// Used in HTML template
func (ht *HTMLReporter) RenderStr(what string) (string, error) {
	var val string
	var err error
	switch what {
	case "showClassStatus":
		val = "show"
		if ht.NoFailure() {
			val = ""
		}
	case "textColorStatus":
		val = "text-danger"
		if ht.NoFailure() {
			val = "text-success"
		}
	case "thumbsStatus":
		val = "down"
		if ht.NoFailure() {
			val = "up"
		}
	case "msgStatus":
		completionStatus := "Experiment completed."
		if !ht.Completed() {
			completionStatus = "Experiment has not completed."
		}
		failureStatus := "Experiment has failures."
		if ht.NoFailure() {
			failureStatus = "Experiment has no failures."
		}
		taskStatus := fmt.Sprintf("%v out of %v tasks are complete.", ht.Result.NumCompletedTasks, len(ht.Spec))
		loopStatus := fmt.Sprintf("%d loops have completed.", ht.Result.NumLoops)
		val = fmt.Sprint(completionStatus)
		val += " "
		val += fmt.Sprint(failureStatus)
		val += " "
		val += fmt.Sprint(taskStatus)
		val += " "
		val += fmt.Sprint(loopStatus)
	default:
		err = fmt.Errorf("do not know how to render %v", what)
	}
	return val, err
}

// MetricDescriptionHTML is used to described metrics in the metrics and SLO section of the HTML report
func (ht *HTMLReporter) MetricDescriptionHTML(metricName string) (string, error) {
	in := ht.Result.Insights
	nm, err := base.NormalizeMetricName(metricName)
	if err != nil {
		return "", err
	}

	m, err := in.GetMetricsInfo(nm)
	if err != nil {
		e := fmt.Errorf("unable to get metrics info for %v", nm)
		log.Logger.Error(e)
		return "", e
	}
	return m.Description, nil
}

// renderSLOSatisfiedHTML provides the HTML icon indicating if the SLO is satisfied
func renderSLOSatisfiedHTML(s bool) string {
	if s {
		return "fa-check-circle"
	}
	return "fa-times-circle"
}

// renderSLOSatisfiedCellClass dictates the cell color indicating if the SLO is satisfied
func renderSLOSatisfiedCellClass(s bool) string {
	if s {
		return "text-success"
	}
	return "text-danger"
}

// SortedVectorMetrics extracts vector metric names from experiment in sorted order
func (ht *HTMLReporter) SortedVectorMetrics() []string {
	keys := []string{}
	for k, mm := range ht.Result.Insights.MetricsInfo {
		if mm.Type == base.HistogramMetricType || mm.Type == base.SampleMetricType {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// sampleHist samples values from a histogram
func sampleHist(h []base.HistBucket) []float64 {
	vals := []float64{}
	for _, b := range h {
		for i := 0; i < int(b.Count); i++ {
			/* #nosec */
			vals = append(vals, b.Lower+(b.Upper-b.Lower)*rand.Float64())
		}
	}
	return vals
}

// VectorMetricValue gets the value of the given vector metric for the given version
// If it is a histogram metric, then its values are sampled from the histogram
// Recall: VectorMetric can be a histogram metric or a sample metric.
func (ht *HTMLReporter) VectorMetricValue(i int, m string) []float64 {
	in := ht.Result.Insights
	mm, ok := in.MetricsInfo[m]
	if !ok {
		log.Logger.Error("could not find vector metric: ", m)
		return nil
	}
	if mm.Type == base.SampleMetricType {
		return in.NonHistMetricValues[i][m]
	}
	// this is a hist metric
	return sampleHist(in.HistMetricValues[i][m])
}

func (ht *HTMLReporter) BestVersions() []string {
	metrics := ht.SortedScalarAndSLOMetrics()
	in := ht.Result.Insights

	results := make([]string, len(metrics))

	rewards := in.Rewards
	winners := in.RewardsWinners

	for i, mn := range metrics {
		j := indexString(rewards.Max, mn)
		if j >= 0 {
			if winners.Max[j] == -1 {
				results[i] = "insufficient data"
			} else {
				results[i] = in.TrackVersionStr(winners.Max[j])
			}
		} else {
			j = indexString(rewards.Min, mn)
			if j >= 0 {
				if winners.Min[j] == -1 {
					results[i] = "insufficient data"
				} else {
					results[i] = in.TrackVersionStr(winners.Min[j])
				}
			} else {
				results[i] = "n/a"
			}
		}
	}

	return results
}
