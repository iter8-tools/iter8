package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/iter8-tools/iter8/base"
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
	XAxisLabel string `json:"xAxisLabel" yaml:"xAxisLabel"`
	Datum      []hist `json:"datum" yaml:"datum"`
}

func (hd *histograms) toJSON() string {
	if hd == nil {
		return ``
	}
	jb, _ := json.Marshal(hd)
	return string(jb)
}

// formatHTML is the HTML template of the experiment results
var formatHTML = `
	<!doctype html>
	<html lang="en">

		{{ headSection }}

		<body>

			{{ dependencies }}

			<div class="container">

				<h1 class="display-4">Experiment Report</h1>
				<h3 class="display-6 text-muted">Status, Insights, and Metrics from Iter8 experiment</h3>
				<hr>

				{{ .HTMLStatus }}

				{{ if .ContainsInsight "SLOs" }} 
					{{ .HTMLSLOSection }}
				{{- end }}

				{{ if .ContainsInsight "HistMetrics" }}
					{{ .HTMLHistMetricsSection }}
				{{- end }}

				{{ if .ContainsInsight "Metrics" }} 
					{{ .HTMLMetricsSection }}
				{{- end }}

			</div>
		
			{{ if .ContainsInsight "HistMetrics" }}
				{{ styleSection }}
				{{ .HTMLHistData }}
				{{ .HTMLHistCharts }}
			{{- end }}

		</body>
	</html>
	`

// styleSection is the style section required for histogram charts
func styleSection() string {
	return `
<style>
.nvd3 text {
	font-size: 16px;
}
svg {
		display: block;
		margin: 0px;
		padding: 0px;
		height: 100%;
		width: 100%;
}
</style>
`
}

// headSection is the fixed head section for experiment report
func headSection() string {
	return `
	<head>
		<!-- Required meta tags -->
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

		<!-- Bootstrap CSS -->
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">

		<style>
			html {
				font-size: 18px;
			}		
		</style>

		<title>Iter8 Experiment Report</title>
	</head>
`
}

// dependencies is the dependencies section for the HTML report
func dependencies() string {
	return `
	<!-- jQuery first, then Popper.js, then Bootstrap JS -->
	<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>

	<!-- NVD3 -->
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.css">
	<!-- Include d3.js first -->
	<script src="https://cdn.jsdelivr.net/npm/d3@3.5.3/d3.min.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.js"></script>
`
}

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
	gramsList := []histograms{}
	for mname, minfo := range e.Result.Insights.MetricsInfo {
		if minfo.Type == base.HistogramMetricType {
			grams := histograms{
				XAxisLabel: fmt.Sprintf("%v", mname),
				Datum:      []hist{},
			}
			for i := 0; i < e.Result.Insights.NumVersions; i++ {
				gram := hist{
					Values: []histBar{},
					Key:    fmt.Sprintf("Version %v", i),
				}
				if counts, ok := e.Result.Insights.MetricValues[i][mname]; ok && len(counts) > 0 {
					for j := 0; j < len(counts); j++ {
						gram.Values = append(gram.Values, histBar{
							X: *minfo.XMin + float64(j)*(*minfo.XMax-*minfo.XMin)/float64(*minfo.NumBuckets),
							Y: counts[j],
						})
					}
					grams.Datum = append(grams.Datum, gram)
				}
			}
			if len(grams.Datum) > 0 {
				gramsList = append(gramsList, grams)
			}
		}
	}
	return gramsList
}

// HTMLHistCharts returns histogram charts section in HTML report
func (e *Experiment) HTMLHistCharts() string {
	return `
	<script>
		var charts = [];
		for (let i = 0; i < chartData.length; i++) {
			nv.addGraph(function() {
				charts.push(nv.models.multiBarChart().stacked(false).showControls(false));
				charts[i]
						.margin({left: 100, bottom: 100})
						.useInteractiveGuideline(true)
						.duration(250)
						;

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

// HTMLStatus prints the current state of the experiment
func (e *Experiment) HTMLStatus() string {
	completionStatus := "Experiment is complete."
	if !e.Completed() {
		completionStatus = "Experiment is yet to complete."
	}

	failureStatus := "Experiment has failures."
	textColor := "text-failure"
	if e.NoFailure() {
		failureStatus = "Experiment has no failures."
		textColor = "text-success"
	}

	taskStatus := fmt.Sprintf("%v out of %v tasks are complete.", len(e.Tasks), e.Result.NumCompletedTasks)

	msg := fmt.Sprintln(taskStatus)
	msg += fmt.Sprintln(failureStatus)
	msg += fmt.Sprintln(completionStatus)

	return fmt.Sprintf(`
	<section class="mt-5">
		<div class="row">
			<div class="col-sm-12">
				<div class="card">
					<h5 class="card-header">Status</h5>
					<div class="card-body">
						<p class="card-text %v">%v</p>
					</div>
				</div>
			</div>
		</div>
	</section>
	`, textColor, msg)
}

// HTMLSLOSection prints the SLO section in HTML report
func (e *Experiment) HTMLSLOSection() string {
	if e.ContainsInsight(base.InsightTypeSLO) {
		if e.printableSLOs() {
			return e.printHTMLSLOs()
		} else {
			return e.printHTMLNoSLOs()
		}
	}
	return ""
}

func (e *Experiment) printHTMLSLOVersions() string {
	in := e.Result.Insights
	out := ""
	if in.NumVersions > 1 {
		for i := 0; i < in.NumVersions; i++ {
			out += fmt.Sprintf(`
			<th scope="col">Version %v</th>
			`, i)
		}
	} else {
		out += `
		<th scope="col">Satisfied</th>
		`
	}
	return out
}

func (e *Experiment) printHTMLSLORows() string {
	in := e.Result.Insights
	out := ""
	for i := 0; i < len(in.SLOStrs); i++ {
		out += `<tr scope="row">` + "\n"
		out += fmt.Sprintf(`
		<td>%v</td>
		`, in.SLOStrs[i])

		for j := 0; j < in.NumVersions; j++ {
			cellClass := "text-success"
			if !in.SLOsSatisfied[i][j] {
				cellClass = "text-danger"
			}
			out += fmt.Sprintf(`
			<td class="%v">%v</td>
			`, cellClass, in.SLOsSatisfied[i][j])
		}
	}
	return out
}

// print HTML SLO validation results
func (e *Experiment) printHTMLSLOs() string {
	sloStrs := `
	<section class="mt-5">
			<h3 class="display-6">Service level objectives (SLOs)</h3>
			<h4 class="display-7 text-muted">Whether or not SLOs are satisfied</h4>
			<hr>
			<table class="table">
			<thead class="thead-light">
				<tr>
					<th scope="col">SLO</th>
	` +
		e.printHTMLSLOVersions() +
		`</tr>
	</thead>
	` +
		`
	<tbody>
	` +
		e.printHTMLSLORows() +
		`
		</tbody>
		</table>
		</section>
		`

	return sloStrs
}

// print HTML no SLOs
func (e *Experiment) printHTMLNoSLOs() string {
	return `
	<section class="mt-5">
		<h2>SLOs Unavailable</h2>
	</section>
	`
}

// HTMLHistMetricsSection prints histogram metrics in the HTML report
func (e *Experiment) HTMLHistMetricsSection() string {
	hd := e.HistData()
	if len(hd) > 0 {
		divs := []string{}
		for i := 0; i < len(hd); i++ {
			divs = append(divs, fmt.Sprintf(`
					<div class='with-3d-shadow with-transitions'>
						<svg id="hist-chart-%v" style="height:500px"></svg>
					</div>
			</section>
			`, i))
		}
		return `
		<section class="mt-5">
		<h3 class="display-6">Histogram Metrics</h3>
		<h4 class="display-7 text-muted">Visualizations for Histogram-type Metrics</h4>
		<hr>

		` + strings.Join(divs, "\n") +
			`</section>`
	}
	return ``
}

func (e *Experiment) printHTMLMetricVersions() string {
	in := e.Result.Insights
	out := ""
	if in.NumVersions > 1 {
		for i := 0; i < in.NumVersions; i++ {
			out += fmt.Sprintf(`
			<th scope="col">Version %v</th>
			`, i)
		}
	} else {
		out += `
		<th scope="col">Metric values</th>
		`
	}
	return out
}

func (e *Experiment) printHTMLMetricRows() string {
	in := e.Result.Insights

	// sort metrics
	keys := []string{}
	for k := range in.MetricsInfo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := ""
	for i := 0; i < len(keys); i++ {
		if e.Result.Insights.MetricsInfo[keys[i]].Type != base.HistogramMetricType {
			u := ""
			// add units if available
			units := e.Result.Insights.MetricsInfo[keys[i]].Units
			if units != nil {
				u += " (" + *units + ")"
			}

			out += `<tr scope="row">` + "\n"
			out += fmt.Sprintf(`
			<td>%v</td>
			`, keys[i]+u)

			for j := 0; j < in.NumVersions; j++ {
				out += fmt.Sprintf(`
				<td>%v</td>
				`, e.getMetricValue(keys[i], j))
			}
		}
	}
	return out
}

// HTMLMetricsSection prints metrics in the HTML report
func (e *Experiment) HTMLMetricsSection() string {
	metricStrs := `
	<section class="mt-5">
			<h3 class="display-6">Metrics</h3>
			<h4 class="display-7 text-muted">Latest observed values of metrics</h4>
			<hr>

			<table class="table">
			<thead class="thead-light">
				<tr>
					<th scope="col">Metrics</th>
	` +
		e.printHTMLMetricVersions() +
		`</tr>
	</thead>
	` +
		`
	<tbody>
	` +
		e.printHTMLMetricRows() +
		`
		</tbody>
		</table>
		</section>
		`

	return metricStrs
}
