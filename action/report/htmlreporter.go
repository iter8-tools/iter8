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

// HTMLHistCharts returns histogram charts section in HTML report
func (r *HTMLReporter) HTMLHistCharts() string {
	return `
	<script>
		var charts = [];
		for (let i = 0; i < chartData.length; i++) {
			nv.addGraph(function() {
				charts.push(nv.models.multiBarChart());

				charts[i]
						.stacked(false)
						.showControls(false)
						.margin({left: 100, bottom: 100})
						.useInteractiveGuideline(true)
						.duration(250);

				var contentGenerator = charts[i].interactiveLayer.tooltip.contentGenerator();
				var tooltip = charts[i].interactiveLayer.tooltip;
				tooltip.headerFormatter(function (d) {
					var lower = d;
					var upper = d + chartData[i].width;
					lower = lower.toFixed(2);
					upper = upper.toFixed(2);
					return "<p>Range: [" + lower + ", " + upper + "]";
				});
				
				// chart sub-models (ie. xAxis, yAxis, etc) when accessed directly, return themselves, not the parent chart, so need to chain separately
				charts[i].xAxis
						.axisLabel(chartData[i].xAxisLabel)
						.tickFormat(function(d) { return d3.format(',.2f')(d);});

				charts[i].yAxis
						.axisLabel('Count')
						.tickFormat(d3.format(',.1f'));

				charts[i].showXAxis(true);

				d3.select('#hist-chart-' + i)
				.datum(chartData[i].datum)
				.transition()
				.call(charts[i]);
						
				nv.utils.windowResize(charts[i].update);
				charts[i].dispatch.on('stateChange', function(e) { nv.log('New State:', JSON.stringify(e)); });
				return charts[i];
		});
	}
	</script>	
	`
}

// RenderStr is a helper method for rendering strings
// Used in HTML template
func (r *HTMLReporter) RenderStr(what string) (string, error) {
	var val string = ""
	var err error = nil
	switch what {
	case "showClassStatus":
		val = "show"
		if r.NoFailure() {
			val = ""
		}
	case "textColorStatus":
		val = "text-danger"
		if r.NoFailure() {
			val = "text-success"
		}
	case "thumbsStatus":
		val = "down"
		if r.NoFailure() {
			val = "up"
		}
	case "msgStatus":
		val = ""
		completionStatus := "Experiment completed."
		if !r.Completed() {
			completionStatus = "Experiment has not completed."
		}
		failureStatus := "Experiment has failures."
		if r.NoFailure() {
			failureStatus = "Experiment has no failures."
		}
		taskStatus := fmt.Sprintf("%v out of %v tasks are complete.", len(r.Tasks), r.Result.NumCompletedTasks)
		val = fmt.Sprint(completionStatus)
		val += " "
		val += fmt.Sprint(failureStatus)
		val += " "
		val += fmt.Sprint(taskStatus)
	default:
		err = fmt.Errorf("do not know how to render %v", what)
	}
	return val, err
}

// MetricDescriptionHTML is used to described metrics in the metrics and SLO section of the HTML report
func (r *HTMLReporter) MetricDescriptionHTML(metricName string) (string, error) {
	in := r.Result.Insights
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
	} else {
		return "fa-times-circle"
	}
}

// renderSLOSatisfiedCellClass dictates the cell color indicating if the SLO is satisfied
func renderSLOSatisfiedCellClass(s bool) string {
	if s {
		return "text-success"
	} else {
		return "text-danger"
	}
}

// SortedVectorMetrics extracts vector metric names from experiment in sorted order
func (r *HTMLReporter) SortedVectorMetrics() []string {
	keys := []string{}
	for k, mm := range r.Result.Insights.MetricsInfo {
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
			vals = append(vals, b.Lower+(b.Upper-b.Lower)*rand.Float64())
		}
	}
	return vals
}

// VectorMetricValue gets the value of the given vector metric for the given version
// If it is a histogram metric, then its values are sampled from the histogram
// Recall: VectorMetric can be a histogram metric or a sample metric.
func (r *HTMLReporter) VectorMetricValue(i int, m string) []float64 {
	in := r.Result.Insights
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
