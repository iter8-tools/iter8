package cmd

import (
	"fmt"
	"sort"

	"github.com/iter8-tools/iter8/base"
)

// formatHTML is the HTML template of the experiment results
var formatHTML = `
	<!doctype html>
	<html lang="en">
		<head>
			<!-- Required meta tags -->
			<meta charset="utf-8">
			<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	
			<!-- Bootstrap CSS -->
			<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
	
			<title>Iter8 Experiment Result</title>
		</head>
		<body>

			<div class="container">
				<h1>Iter8 Experiment Report</h1>
				<hr>

				{{ .HTMLState }}

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
		
			<!-- jQuery first, then Popper.js, then Bootstrap JS -->
			<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
			<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>

			<!-- NVD3 -->
		  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.css">
			<!-- Include d3.js first -->
			<script src="https://cdn.jsdelivr.net/npm/d3@3.5.3/d3.min.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.js"></script>

			<style>
				text {
						font: 12px sans-serif;
				}
				svg {
						display: block;
						margin: 0px;
						padding: 0px;
						height: 100%;
						width: 100%;
				}
			</style>

			{{ if .ContainsInsight "HistMetrics" }}
				<script>

				var chart;
				nv.addGraph(function() {
						chart = nv.models.multiBarChart().stacked(false).showControls(false);
						chart
								.margin({left: 100, bottom: 100})
								.useInteractiveGuideline(true)
								.duration(250)
								;
		
						// chart sub-models (ie. xAxis, yAxis, etc) when accessed directly, return themselves, not the parent chart, so need to chain separately
						chart.xAxis
								.axisLabel("Latency (msec)")
								.tickFormat(function(d) { return d3.format(',.2f')(d);});
		
						chart.yAxis
								.axisLabel('Count')
								.tickFormat(d3.format(',.1f'));
		
						chart.showXAxis(true);
		
						d3.select('#hist-chart-1')
								.datum(sinAndCos())
								.transition()
								.call(chart);

						d3.select('#hist-chart-2')
						.datum(sinData())
						.transition()
						.call(chart);
								
						nv.utils.windowResize(chart.update);
						chart.dispatch.on('stateChange', function(e) { nv.log('New State:', JSON.stringify(e)); });
						return chart;
				});
		
				//Simple test data generators
				function sinAndCos() {
						var sin = [],
								cos = [];
		
						for (var i = 0; i < 30; i++) {
								sin.push({x: i, y: Math.sin(i/10)});
								cos.push({x: i, y: Math.cos(i/10)});
						}
		
						return [
							{values: sin, key: "Version 0"},
							{values: cos, key: "Version 1"}
						];
				}
		
				function sinData() {
						var sin = [];
		
						for (var i = 0; i < 10; i++) {
								sin.push({x: i, y: Math.sin(i/10) * Math.random() * 100});
						}
		
						return [{
								values: sin,
								key: "Version 0",
						}];
				}
		
				</script>			
			{{- end }}

		</body>
	</html>
	`

// HTMLState prints the current state of the experiment
func (e *Experiment) HTMLState() string {
	return fmt.Sprintf(`
	<section>
		<h2>Summary</h2>
		<ul class="list-group">
			<li class="list-group-item d-flex justify-content-between align-items-center">
				Experiment completed
				<span><strong>%v</strong></span>
			</li>
			<li class="list-group-item d-flex justify-content-between align-items-center">
				Experiment failed
				<span><strong>%v</strong></span>
			</li>
			<li class="list-group-item d-flex justify-content-between align-items-center">
				Number of completed tasks
				<span><strong>%v</strong></span>
			</li>
		</ul>
	</section>
	<hr>`, e.Completed(), !e.NoFailure(), len(e.tasks))
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
	if *in.NumAppVersions > 1 {
		for i := 0; i < *in.NumAppVersions; i++ {
			out += fmt.Sprintf(`
			<th scope="col">Version %v</th>
			`, i)
		}
	} else {
		out += `
		<th scope="col">SLO satisfied</th>
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

		for j := 0; j < *in.NumAppVersions; j++ {
			out += fmt.Sprintf(`
			<td>%v</td>
			`, in.SLOsSatisfied[i][j])
		}
	}
	return out
}

// print HTML SLO validation results
func (e *Experiment) printHTMLSLOs() string {
	sloStrs := `
	<section>
			<h2>Service level objectives (SLOs)</h2>
			<p>Whether or not SLOs are satisfied</p>
			<table class="table">
			<thead class="thead-dark">
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
		<hr>
		`

	return sloStrs
}

// print HTML no SLOs
func (e *Experiment) printHTMLNoSLOs() string {
	return `
	<section>
		<h2>SLOs Unavailable</h2>
	</section>
	<hr>`
}

// HTMLHistMetricsSection prints histogram metrics in the HTML report
func (e *Experiment) HTMLHistMetricsSection() string {
	return `
	<section>
		<h2>Histogram Metrics</h2>
			<div class='with-3d-shadow with-transitions'>
				<svg id="hist-chart-1" style="height:500px"></svg>
			</div>
			<div class='with-3d-shadow with-transitions'>
				<svg id="hist-chart-2" style="height:500px"></svg>
			</div>
	</section>
	<hr>`
}

func (e *Experiment) printHTMLMetricVersions() string {
	in := e.Result.Insights
	out := ""
	if *in.NumAppVersions > 1 {
		for i := 0; i < *in.NumAppVersions; i++ {
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

		for j := 0; j < *in.NumAppVersions; j++ {
			out += fmt.Sprintf(`
			<td>%v</td>
			`, e.getMetricValue(keys[i], j))
		}
	}
	return out
}

// HTMLMetricsSection prints metrics in the HTML report
func (e *Experiment) HTMLMetricsSection() string {
	metricStrs := `
	<section>
			<h2>Metrics</h2>
			<p>Latest observed values of metrics</p>
			<table class="table">
			<thead class="thead-dark">
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
		<hr>
		`

	return metricStrs
}

// HTMLHistMetricsDataSection prints the data needed for
// the Hist Metrics section in HTML report
func (e *Experiment) HTMLHistMetricsDataSection() string {
	return ""
}
