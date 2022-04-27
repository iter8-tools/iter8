{{- define "metrics.istio" -}}
url: {{ .Values.Endpoint }}/api/v1/query
provider: Istio
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
#   destination_workload_namespacee           string
#   StartingTime                 int64 (UNIX time stamp)
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
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
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
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
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
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
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
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_request_duration_milliseconds_bucket{
        le='+Inf',
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
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
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        {{- if .Values.reporter }}
          reporter="{{.Values.reporter}}",
        {{- end }}
        {{- if .Values.app }}
          app="{{.Values.app}}",
        {{- end }}
        {{- if .Values.chart }}
          chart="{{.Values.chart}}",
        {{- end }}
        {{- if .Values.connection_security_policy }}
          connection_security_policy="{{.Values.connection_security_policy}}",
        {{- end }}
        {{- if .Values.destination_app }}
          destination_app="{{.Values.destination_app}}",
        {{- end }}
        {{- if .Values.destination_canonical_revision }}
          destination_canonical_revision="{{.Values.destination_canonical_revision}}",
        {{- end }}
        {{- if .Values.destination_canonical_service }}
          destination_canonical_service="{{.Values.destination_canonical_service}}",
        {{- end }}
        {{- if .Values.destination_cluster }}
          destination_cluster="{{.Values.destination_cluster}}",
        {{- end }}
        {{- if .Values.destination_principal }}
          destination_principal="{{.Values.destination_principal}}",
        {{- end }}
        {{- if .Values.destination_service }}
          destination_service="{{.Values.destination_service}}",
        {{- end }}
        {{- if .Values.destination_service_name }}
          destination_service_name="{{.Values.destination_service_name}}",
        {{- end }}
        {{- if .Values.destination_service_namespace }}
          destination_service_namespace="{{.Values.destination_service_namespace}}",
        {{- end }}
        {{- if .Values.destination_version }}
          destination_version="{{.Values.destination_version}}",
        {{- end }}
        {{- if .Values.heritage }}
          heritage="{{.Values.heritage}}",
        {{- end }}
        {{- if .Values.install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.Values.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .Values.instance }}
          instance="{{.Values.instance}}",
        {{- end }}
        {{- if .Values.istio }}
          istio="{{.Values.istio}}",
        {{- end }}
        {{- if .Values.istio_io_rev }}
          istio_io_rev="{{.Values.istio_io_rev}}",
        {{- end }}
        {{- if .Values.job }}
          job="{{.Values.job}}",
        {{- end }}
        {{- if .Values.namespace }}
          namespace="{{.Values.namespace}}",
        {{- end }}
        {{- if .Values.operator_istio_io_component }}
          operator_istio_io_component="{{.Values.operator_istio_io_component}}",
        {{- end }}
        {{- if .Values.pod }}
          pod="{{.Values.pod}}",
        {{- end }}
        {{- if .Values.pod_template_hash }}
          pod_template_hash="{{.Values.pod_template_hash}}",
        {{- end }}
        {{- if .Values.release }}
          release="{{.Values.release}}",
        {{- end }}
        {{- if .Values.request_protocol }}
          request_protocol="{{.Values.request_protocol}}",
        {{- end }}
        {{- if .Values.response_code }}
          response_code="{{.Values.response_code}}",
        {{- end }}
        {{- if .Values.response_flags }}
          response_flags="{{.Values.response_flags}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.Values.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .Values.service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.Values.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .Values.sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.Values.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .Values.source_app }}
          source_app="{{.Values.source_app}}",
        {{- end }}
        {{- if .Values.source_canonical_revision }}
          source_canonical_revision="{{.Values.source_canonical_revision}}",
        {{- end }}
        {{- if .Values.source_canonical_service }}
          source_canonical_service="{{.Values.source_canonical_service}}",
        {{- end }}
        {{- if .Values.source_cluster }}
          source_cluster="{{.Values.source_cluster}}",
        {{- end }}
        {{- if .Values.source_principal }}
          source_principal="{{.Values.source_principal}}",
        {{- end }}
        {{- if .Values.source_version }}
          source_version="{{.Values.source_version}}",
        {{- end }}
        {{- if .Values.source_workload }}
          source_workload="{{.Values.source_workload}}",
        {{- end }}
        {{- if .Values.source_workload_namespace }}
          source_workload_namespace="{{.Values.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]
{{- end }}