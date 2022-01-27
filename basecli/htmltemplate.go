package basecli

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

// var templateHTML = `
// 	<!doctype html>
// 	<html lang="en">

// 		<head>
// 			<!-- Required meta tags -->
// 			<meta charset="utf-8">
// 			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

// 			<!-- Font Awesome -->
// 			<script src="https://kit.fontawesome.com/db794f5235.js" crossorigin="anonymous"></script>

// 			<!-- Bootstrap CSS -->
// 			<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">

// 			<style>
// 				html {
// 					font-size: 18px;
// 				}
// 			</style>

// 			<title>Iter8 Experiment Report</title>
// 		</head>

// 		<body>

// 			<!-- jQuery first, then Popper.js, then Bootstrap JS -->
// 			<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
// 			<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
// 			<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>

// 			<!-- NVD3 -->
// 		  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.css">
// 			<!-- Include d3.js first -->
// 			<script src="https://cdn.jsdelivr.net/npm/d3@3.5.3/d3.min.js"></script>
// 			<script src="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.js"></script>

// 			<div class="container">

// 				<h1 class="display-4">Experiment Report</h1>
// 				<h3 class="display-6">Insights from Iter8 Experiment</h3>
// 				<hr>

// 				{{ .HTMLStatus }}

// 				{{ if .ContainsInsight "SLOs" }}
// 					{{ .HTMLSLOSection }}
// 				{{- end }}

// 				{{ if .ContainsInsight "HistMetrics" }}
// 					{{ .HTMLHistMetricsSection }}
// 				{{- end }}

// 				{{ if .ContainsInsight "Metrics" }}
// 					{{ .HTMLMetricsSection }}
// 				{{- end }}

// 			</div>

// 			{{ if .ContainsInsight "HistMetrics" }}
// 				<style>
// 					.nvd3 text {
// 						font-size: 16px;
// 					}
// 					svg {
// 							display: block;
// 							margin: 0px;
// 							padding: 0px;
// 							height: 100%;
// 							width: 100%;
// 					}
// 				</style>
// 				{{ .HTMLHistData }}
// 				{{ .HTMLHistCharts }}
// 			{{- end }}

// 		</body>
// 	</html>
// 	`

// HTMLHistData returns histogram data section in HTML report
func (e *Experiment) HTMLHistData() string {
	hds := e.HistData()
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
func (e *Experiment) HistData() []histograms {
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
func (e *Experiment) HTMLHistCharts() string {
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
func (e *Experiment) RenderStr(what string) (string, error) {
	var val string = ""
	var err error = nil
	switch what {
	case "showClassStatus":
		val = "show"
		if e.NoFailure() {
			val = ""
		}
	case "textColorStatus":
		val = "text-danger"
		if e.NoFailure() {
			val = "text-success"
		}
	case "thumbsStatus":
		val = "down"
		if e.NoFailure() {
			val = "up"
		}
	case "msgStatus":
		val = ""
		completionStatus := "Experiment completed."
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

func (e *Experiment) MetricWithUnits(metricName string) (string, error) {
	in := e.Result.Insights
	nm, err := base.NormalizeMetricName(metricName)
	if err != nil {
		return "", err
	}
	m, ok := in.MetricsInfo[nm]
	if !ok {
		e := fmt.Errorf("unknown metric name %v", nm)
		log.Logger.Error(e)
		return "", e
	}
	str := nm
	if m.Units != nil {
		str = fmt.Sprintf("%v (%v)", str, *m.Units)
	}
	return str, nil
}

func (e *Experiment) MetricDescriptionHTML(metricName string) (string, error) {
	in := e.Result.Insights
	nm, err := base.NormalizeMetricName(metricName)
	if err != nil {
		return "", err
	}
	m, ok := in.MetricsInfo[nm]
	if !ok {
		e := fmt.Errorf("unknown metric name %v", nm)
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
func (e *Experiment) SortedVectorMetrics() []string {
	keys := []string{}
	for k, mm := range e.Result.Insights.MetricsInfo {
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
func (e *Experiment) VectorMetricValue(i int, m string) []float64 {
	in := e.Result.Insights
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
