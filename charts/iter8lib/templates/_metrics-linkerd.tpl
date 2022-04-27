{{- define "metrics.linkerd" -}}
url: {{ .Values.Endpoint }}/api/v1/query
provider: Linkerd
method: GET
# Inputs for the template:
#
# Inputs for the metrics (output of template):
#   deployment           string
#   namespace            string
#   StartingTime         int64 (UNIX time stamp)
#
# Note: ElapsedTime is produced by Iter8
metrics:
- name: request-count
  type: counter
  description: |
    Number of requests
  params:
  - name: query
    value: |
      sum(last_over_time(request_total{
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
- name: error-count
  type: counter
  description: |
    Number of non-successful requests
  params:
  - name: query
    value: |
      sum(last_over_time(response_total{
        response_code=~'5..',
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
- name: error-rate
  type: gauge
  description: |
    Percentage of non-successful requests
  params:
  - name: query
    value: |
      sum(last_over_time(response_total{
        response_code=~'5..',
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(response_total{
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result.[0].value.[1]
- name: le500ms-latency-percentile
  type: gauge
  description: |
    Less than 500 ms latency
  params:
  - name: query
    value: |
      sum(last_over_time(response_latency_ms_bucket{
        le='500',
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(response_latency_ms_bucket{
        le='+Inf',
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
- name: mean-latency
  type: gauge
  description: |
    Mean latency
  params:
  - name: query
    value: |
      sum(last_over_time(response_latency_ms_sum{
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(request_total{
        direction='inbound',
        tls='true',
        {{- include "metrics.common.linkerd" . }}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
{{- end }}