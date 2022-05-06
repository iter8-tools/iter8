{{- define "metrics.istio" -}}
url: {{ .Values.endpoint }}/api/v1/query
provider: istio
method: GET
# Inputs for the template:
#   app                                       string
#   chart                                     string
#   connection_security_policy                string
#   destination_app                           string
#   destination_canonical_revision            string
#   destination_canonical_service             string
#   destination_cluster                       string
#   destination_principal                     string
#   destination_service                       string
#   destination_service_name                  string
#   destination_service_namespace             string
#   destination_version                       string
#   heritage                                  string
#   install_operator_istio_io_owning_resource string
#   instance                                  string
#   istio                                     string
#   istio_io_rev                              string
#   job                                       string
#   namespace                                 string
#   operator_istio_io_component               string
#   pod                                       string
#   pod_template_hash                         string
#   release                                   string
#   request_protocol                          string
#   response_code                             string
#   response_flags                            string
#   service_istio_io_canonical_name           string
#   service_istio_io_canonical_revision       string
#   sidecar_istio_io_inject                   string
#   source_app                                string
#   source_canonical_revision                 string
#   source_canonical_service                  string
#   source_cluster                            string
#   source_principal                          string
#   source_version                            string
#   source_workload                           string
#   source_workload_namespace                 string
#
# Inputs for the metrics (output of template):
#   destination_workload                      string
#   destination_workload_namespace            string
#   StartingTime                              string
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
      sum(last_over_time(istio_requests_total{
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
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
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
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
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
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
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_request_duration_milliseconds_bucket{
        le='+Inf',
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
- name: mean-latency
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
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- include "metrics.common.istio" . | indent 4 }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
{{- end }}