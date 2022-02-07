<!doctype html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <!-- Font Awesome -->
    <script src="https://kit.fontawesome.com/db794f5235.js" crossorigin="anonymous"></script>
    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
    <style>
      html {
        font-size: 18px;
      }		
    </style>
    <title>Iter8 Experiment Report</title>
  </head>

  <body>
    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
  
    <!-- Plotly.js -->
    <script src="https://cdn.plot.ly/plotly-2.8.3.min.js"></script>

    <div class="container">
      <h1 class="display-4"><a href="https://iter8.tools">Iter8</a> Experiment Report</h1>
      <hr>
      
      <div class="toast fade {{ .RenderStr "showClassStatus" }} mw-100" role="alert" aria-live="assertive" aria-atomic="true">
        <div class="toast-header {{ .RenderStr "textColorStatus" }}">
          <strong class="mr-auto">
            Experiment Status
            &nbsp;&nbsp;
            <i class="fas fa-thumbs-{{ .RenderStr "thumbsStatus" }}"></i>
          </strong>
          <button type="button" class="ml-2 mb-1 close" data-dismiss="toast" aria-label="Close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="toast-body {{ .RenderStr "textColorStatus" }}">
          {{ .RenderStr "msgStatus" }}
        </div>
      </div>
    
      <script>
      $(document).ready(function(){
        $(function () {
          $('[data-toggle="tooltip"]').tooltip()
        });	

        $(".toast").toast({
          autohide: false,
          delay: 10000
        }).toast('show');
      });	
      </script>

      {{- if not (empty .Result.Insights.SLOs) }}  
      <section class="mt-5">
        <h3 class="display-6">Service level objectives (SLOs)</h3>
        <h4 class="display-7 text-muted">Whether or not SLOs are satisfied</h4>
        <hr>
        <table class="table">
          <thead class="thead-light">
            <tr>
              <th scope="col">SLO Conditions</th>
              {{- if ge .Result.Insights.NumVersions 2 }}
              {{- range until .Result.Insights.NumVersions }}
              <th scope="col" class="text-center">Version {{ . }} </th>
              {{- end}}              
              {{- else }} 
              <th scope="col" class="text-center">Satisfied</th>
              {{- end }}
            </tr>
          </thead>
          <tbody>
              {{- range $ind, $slo := .Result.Insights.SLOs }}
              <tr scope="row">
                <td>
                  {{ if $slo.LowerLimit }} {{- $slo.LowerLimit }} &leq; {{ end -}}
                  <a href="javascript:void(0)" data-toggle="tooltip" data-placement="top" title="{{ $.MetricDescriptionHTML $slo.Metric }}">
                    {{ $.MetricWithUnits $slo.Metric }}
                  </a>
                  {{ if $slo.UpperLimit }} &leq; {{ $slo.UpperLimit -}} {{- end }}
                </td>
                {{- range (index $.Result.Insights.SLOsSatisfied $ind) }}
                <td class="{{ renderSLOSatisfiedCellClass .  }} text-center">
                  <i class="far {{ renderSLOSatisfiedHTML . }}"></i>                
                </td>
                {{- end }}
              </tr>
              {{- end}}          
          </tbody>
        </table>
      </section>
      {{- end }}

  		<section class="mt-5">
        <h3 class="display-6">Metric Histograms</h3>
        <hr>

        {{- range $ind, $mn := .SortedVectorMetrics }}
        <div id="vm-{{ $mn }}"></div>
        <script>
          var data = [];
          {{- range until $.Result.Insights.NumVersions }}
          data.push({
            x: {{ $.VectorMetricValue . $mn }},
            name: "Version {{ . }}",
            histnorm: "percent", 
            opacity: 0.5, 
            type: "histogram"
          })
          {{- end }}

          var layout = {
            bargroupgap: 0.2, 
            barmode: "overlay", 
            title: "Histogram of {{ $.MetricWithUnits $mn }}", 
            xaxis: {title: "{{ $.MetricWithUnits $mn }}"}, 
            yaxis: {title: "% of observations"}
          };
          Plotly.newPlot("vm-{{ $mn }}", data, layout);
        </script>
        {{- end }}
			</section>

  	  <section class="mt-5">
        <h3 class="display-6">Latest observed values for metrics</h3>
        <hr>
        <table class="table">
          <thead class="thead-light">
            <tr>
              <th scope="col">Metric</th>
              {{- if ge .Result.Insights.NumVersions 2 }}
              {{- range until .Result.Insights.NumVersions }}
              <th scope="col" class="text-center">Version {{ . }} </th>
              {{- end}}              
              {{- else }} 
              <th scope="col" class="text-center">Value</th>
              {{- end }}
            </tr>
          </thead>
          <tbody>
              {{- range $ind, $mn := .SortedScalarAndSLOMetrics }}
              <tr scope="row">
                <td>
                  <a href="javascript:void(0)" data-toggle="tooltip" data-placement="top" title="{{ $.MetricDescriptionHTML $mn }}">
                    {{ $.MetricWithUnits $mn }}
                  </a>
                </td>
                {{- range until $.Result.Insights.NumVersions }}
                <td class="text-center">
                {{ $.ScalarMetricValueStr . $mn }}
                </td>
                {{- end }}
              </tr>
              {{- end}}          
          </tbody>
        </table>
      </section>

    </div>
  </body>
</html>
