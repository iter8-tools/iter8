package report

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	_ "embed"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
)

type htmlReporter Reporter

type histBar struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
}

type hist struct {
	Values []histBar `json:"values" yaml:"values"`
	Key    string    `json:"key" yaml:"key"`
}

type histograms struct {
	XAxisLabel string  `json:"xAxisLabel" yaml:"xAxisLabel"`
	Datum      []hist  `json:"datum" yaml:"datum"`
	Width      float64 `json:"width" yaml:"width"`
}

func (hd *histograms) toJSON() string {
	if hd == nil {
		return ``
	}
	jb, _ := json.Marshal(hd)
	return string(jb)
}

// reportHTML is the HTML report template
//go:embed htmlreport.tpl
var reportHTML string

// HTMLHistData returns histogram data section in HTML report
func (r *htmlReporter) HTMLHistData() string {
	hds := r.HistData()
	hdsJSONs := []string{}
	for _, hd := range hds {
		hdsJSONs = append(hdsJSONs, hd.toJSON())
	}
	htmlhd := fmt.Sprintf(`
	<script>
		chartData = [%v]
	</script>
	`, strings.Join(hdsJSONs, ", \n"))
	return htmlhd
}

// HistData provides histogram data for all histogram metrics
func (r *htmlReporter) HistData() []histograms {
	return nil
	// gramsList := []histograms{}
	// for mname, minfo := range e.Result.Insights.MetricsInfo {
	// 	if minfo.Type == base.HistogramMetricType {
	// 		// figure out xAxisLabel
	// 		xAxisLabel := fmt.Sprintf("%v", mname)
	// 		if minfo.Units != nil {
	// 			xAxisLabel += " (" + *minfo.Units + ")"
	// 		}

	// 		grams := histograms{
	// 			XAxisLabel: xAxisLabel,
	// 			Datum:      []hist{},
	// 			Width:      (*minfo.XMax - *minfo.XMin) / float64(*minfo.NumBuckets),
	// 		}
	// 		for i := 0; i < e.Result.Insights.NumVersions; i++ {
	// 			key := fmt.Sprintf("Version %v", i)
	// 			if e.Result.Insights.NumVersions == 1 {
	// 				key = "count"
	// 			}
	// 			gram := hist{
	// 				Values: []histBar{},
	// 				Key:    key,
	// 			}
	// 			if counts, ok := e.Result.Insights.MetricValues[i][mname]; ok && len(counts) > 0 {
	// 				for j := 0; j < len(counts); j++ {
	// 					gram.Values = append(gram.Values, histBar{
	// 						X: *minfo.XMin + float64(j)*(*minfo.XMax-*minfo.XMin)/float64(*minfo.NumBuckets),
	// 						Y: counts[j],
	// 					})
	// 				}
	// 				grams.Datum = append(grams.Datum, gram)
	// 			}
	// 		}
	// 		if len(grams.Datum) > 0 {
	// 			gramsList = append(gramsList, grams)
	// 		}
	// 	}
	// }
	// return gramsList
}

// HTMLHistCharts returns histogram charts section in HTML report
func (r *htmlReporter) HTMLHistCharts() string {
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

// RenderStrHTML is a helper method for rendering strings
// Used in HTML template
func (r *htmlReporter) RenderStr(what string) (string, error) {
	var val string = ""
	var err error = nil
	switch what {
	case "showClassStatus":
		val = "show"
		if e := base.Experiment(*r); e.NoFailure() {
			val = ""
		}
	case "textColorStatus":
		val = "text-danger"
		if e := base.Experiment(*r); e.NoFailure() {
			val = "text-success"
		}
	case "thumbsStatus":
		val = "down"
		if e := base.Experiment(*r); e.NoFailure() {
			val = "up"
		}
	case "msgStatus":
		val = ""
		completionStatus := "Experiment completed."
		e := base.Experiment(*r)
		if !e.Completed() {
			completionStatus = "Experiment has not completed."
		}
		failureStatus := "Experiment has failures."
		if e.NoFailure() {
			failureStatus = "Experiment has no failures."
		}
		taskStatus := fmt.Sprintf("%v out of %v tasks are complete.", len(e.Tasks), e.Result.NumCompletedTasks)
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

func (r *htmlReporter) MetricWithUnits(metricName string) (string, error) {
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
	str := nm
	if m.Units != nil {
		str = fmt.Sprintf("%v (%v)", str, *m.Units)
	}
	return str, nil
}

func (r *htmlReporter) MetricDescriptionHTML(metricName string) (string, error) {
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

func renderSLOSatisfiedHTML(s bool) string {
	if s {
		return "fa-check-circle"
	} else {
		return "fa-times-circle"
	}
}

func renderSLOSatisfiedCellClass(s bool) string {
	if s {
		return "text-success"
	} else {
		return "text-danger"
	}
}

// SortedVectorMetrics extracts vector metric names from experiment in sorted order
func (r *htmlReporter) SortedVectorMetrics() []string {
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
func (r *htmlReporter) VectorMetricValue(i int, m string) []float64 {
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
