# This file provides templated metric specifications that enable
# Iter8 to retrieve metrics from Istio's Prometheus add-on.
# 
# For a list of metrics supported out-of-the-box by the Istio Prometheus add-on, 
# please see https://istio.io/latest/docs/reference/config/metrics/
#
# Iter8 substitutes the placeholders in this file with values, 
# and uses the resulting metric specs to query Prometheus.
# The placeholders are as follows.
# 
# labels                          map[string]interface{}              optional
# elapsedTimeSeconds              int                                 implicit
# startingTime                    string                              optional
# latencyPercentiles              []int                               optional
#
# labels: this is the set of Prometheus labels that will be used to identify a particular
# app version. These labels will be applied to every Prometheus query. To learn more
# about what labels you can use for Prometheus, please see
# https://istio.io/latest/docs/reference/config/metrics/#labels
#
# elapsedTimeSeconds: this should not be specified directly by the user. 
# It is implicitly computed by Iter8 according to the following formula
# elapsedTimeSeconds := (time.Now() - startingTime).Seconds()
# 
# startingTime: By default, this is the time at which the Iter8 experiment started.
# The user can explicitly specify the startingTime for each app version
# (for example, the user can set the startingTime to the creation time of the app version)
#
# latencyPercentiles: Each item in this slice will create a new metric spec.
# For example, if this is set to [50,75,90,95],
# then, latency-p50, latency-p75, latency-p90, latency-p95 metric specs are created.

{{- define "labels"}}
{{- range $key, $val := .labels }}
{{- if or (eq (kindOf $val) "slice") (eq (kindOf $val) "map")}}
{{- fail (printf "labels should be a primitive types but received: %s :%s" $key $val) }}
{{- end }}
{{- if eq $key "response_code"}}
{{- fail "labels should not contain 'response_code'" }}
{{- end }}
        {{ $key }}="{{ $val }}",
{{- end }}
{{- end}}

# url is the HTTP endpoint where the Prometheus service installed by Istio's Prom add-on
# can be queried for metrics

url: {{ .istioPromURL | default "http://prometheus.istio-system:9090/api/v1/query" }}
provider: istio-prom
method: GET
metrics:
- name: request-count
  type: counter
  description: |
    Number of requests
  params:
  - name: query
    value: |
      sum(last_over_time(istio_requests_total{
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
- name: error-count
  type: counter
  description: |
    Number of unsuccessful requests
  params:
  - name: query
    value: |
      sum(last_over_time(istio_requests_total{
        response_code=~'5..',
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1] | tonumber
- name: error-rate
  type: gauge
  description: |
    Fraction of unsuccessful requests
  params:
  - name: query
    value: |
      (sum(last_over_time(istio_requests_total{
        response_code=~'5..',
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0))/(sum(last_over_time(istio_requests_total{
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0))
  jqExpression: .data.result.[0].value.[1] | tonumber
- name: latency-mean
  type: gauge
  description: |
    Mean latency
  params:
  - name: query
    value: |
      (sum(last_over_time(istio_request_duration_milliseconds_sum{
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0))/(sum(last_over_time(istio_requests_total{
        {{- template "labels" . }}
      }[{{ .elapsedTimeSeconds }}s])) or on() vector(0))
  jqExpression: .data.result[0].value[1] | tonumber
{{- range $i, $p := .latencyPercentiles }}
- name: latency-p{{ $p }}
  type: gauge
  description: |
    {{ $p }} percentile latency
  params:
  - name: query
    value: |
      histogram_quantile(0.{{ $p }}, sum(rate(istio_request_duration_milliseconds_bucket{
        {{- template "labels" $ }}
      }[{{ $.elapsedTimeSeconds }}s])) by (le))
  jqExpression: .data.result[0].value[1] | tonumber
{{- end }}
