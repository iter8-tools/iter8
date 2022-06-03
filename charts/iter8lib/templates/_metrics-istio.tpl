{{- define "metrics.istio" -}}
url: {{ .Values.endpoint }}/api/v1/query
provider: istio
method: GET
# Inputs for the metrics (output of template):
#   destination_workload                      string
#   destination_workload_namespace            string
#   startingTime                              string
#
# Note: elapsedTimeSeconds is produced by Iter8
metrics:
- name: request-count
  type: counter
  description: |
    Number of requests
  params:
  - name: query
    value: |
      sum(last_over_time(istio_requests_total{
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
- name: error-count
  type: counter
  description: |
    Number of non-successful requests
  params:
  - name: query
    value: |
      sum(last_over_time(istio_requests_total{
        response_code=~'5..',
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
- name: error-rate
  type: gauge
  description: |
    Percentage of non-successful requests
  params:
  - name: query
    value: |
      sum(last_over_time(istio_requests_total{
        response_code=~'5..',
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.response_code }}
        response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result.[0].value.[1]
- name: le500ms-latency-percentile
  type: gauge
  description: |
    Less than 500 ms latency
  params:
  - name: query
    value: |
      sum(last_over_time(istio_request_duration_milliseconds_bucket{
        le='500',
        {{- if .Values.response_code }}
        response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_request_duration_milliseconds_bucket{
        le='+Inf',
        {{- if .Values.response_code }}
        response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
- name: latency-mean
  type: gauge
  description: |
    Mean latency
  params:
  - name: query
    value: |
      sum(last_over_time(istio_request_duration_milliseconds_sum{
        {{- if .Values.response_code }}
        response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.response_code }}
        response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
        reporter="{{.Values.reporter}}",
        {{- end }}
        destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
      }[{{"{{"}}.elapsedTimeSeconds{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
{{- end }}