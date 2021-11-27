package cmd

import (
	"fmt"

	"github.com/iter8-tools/iter8/base"
)

// formatHTML is the HTML template of the experiment results
func formatHTML(e *Experiment) string {
	return `
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
				<hr>` +
		fmt.Sprintln() +
		e.printHTMLState() +
		fmt.Sprintln() +
		e.printHTMLSLOSection() +
		fmt.Sprintln() +
		`

				<div>
					<canvas id="myChart"></canvas>
				</div>

			</div>
		
			<!-- jQuery first, then Popper.js, then Bootstrap JS -->
			<script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
			<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>

			<!-- Chart JS -->
			<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>

			<script>
				const labels = [
					'January',
					'February',
					'March',
					'April',
					'May',
					'June',
				];
				const data = {
					labels: labels,
					datasets: [{
						label: 'My First dataset',
						backgroundColor: 'rgb(255, 99, 132)',
						borderColor: 'rgb(255, 99, 132)',
						data: [0, 10, 5, 2, 20, 30, 45],
					}]
				};
				
				const config = {
					type: 'line',
					data: data,
					options: {}
				};
			
				const myChart = new Chart(
					document.getElementById('myChart'),
					config
				);
			</script>



		</body>
	</html>
	`
}

// print the current state of the experiment
func (e *Experiment) printHTMLState() string {
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

// print the SLO section
func (e *Experiment) printHTMLSLOSection() string {
	if e.containsInsight(base.InsightTypeSLO) {
		if e.printableSLOs() {
			return e.printHTMLSLOs()
		} else {
			return e.printHTMLNoSLOs()
		}
	}
	return ""
}

// print HTML SLO validation results
func (e *Experiment) printHTMLSLOs() string {
	return `
	<section>
			<table class="table">
			<thead class="thead-dark">
				<tr>
					<th scope="col">SLOs</th>
					<th scope="col">First</th>
					<th scope="col">Last</th>
					<th scope="col">Handle</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<th scope="row">1</th>
					<td>Mark</td>
					<td>Otto</td>
					<td>@mdo</td>
				</tr>
				<tr>
					<th scope="row">2</th>
					<td>Jacob</td>
					<td>Thornton</td>
					<td>@fat</td>
				</tr>
				<tr>
					<th scope="row">3</th>
					<td>Larry</td>
					<td>the Bird</td>
					<td>@twitter</td>
				</tr>
			</tbody>
		</table>	
	</section>
	<hr>`
}

// print HTML no SLOs
func (e *Experiment) printHTMLNoSLOs() string {
	return `
	<section>
		<h2>SLOs Unavailable</h2>
	</section>
	<hr>`
}
