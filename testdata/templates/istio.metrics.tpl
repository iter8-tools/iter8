url: {{ .Endpoint }}/api/v1/query
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
#   destination_workload_namespacee            string
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
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
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
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
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
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
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
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_request_duration_milliseconds_bucket{
        le='+Inf',
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
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
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(istio_requests_total{
        reporter='source',
        {{- if .app }}
          app="{{.app}}",
        {{- end }}
        {{- if .chart }}
          chart="{{.chart}}",
        {{- end }}
        {{- if .connection_security_policy }}
          connection_security_policy="{{.connection_security_policy}}",
        {{- end }}
        {{- if .destination_app }}
          destination_app="{{.destination_app}}",
        {{- end }}
        {{- if .destination_canonical_revision }}
          destination_canonical_revision="{{.destination_canonical_revision}}",
        {{- end }}
        {{- if .destination_canonical_service }}
          destination_canonical_service="{{.destination_canonical_service}}",
        {{- end }}
        {{- if .destination_cluster }}
          destination_cluster="{{.destination_cluster}}",
        {{- end }}
        {{- if .destination_principal }}
          destination_principal="{{.destination_principal}}",
        {{- end }}
        {{- if .destination_service }}
          destination_service="{{.destination_service}}",
        {{- end }}
        {{- if .destination_service_name }}
          destination_service_name="{{.destination_service_name}}",
        {{- end }}
        {{- if .destination_service_namespace }}
          destination_service_namespace="{{.destination_service_namespace}}",
        {{- end }}
        {{- if .destination_version }}
          destination_version="{{.destination_version}}",
        {{- end }}
        {{- if .heritage }}
          heritage="{{.heritage}}",
        {{- end }}
        {{- if .install_operator_istio_io_owning_resource }}
          install_operator_istio_io_owning_resource="{{.install_operator_istio_io_owning_resource}}",
        {{- end }}
        {{- if .instance }}
          instance="{{.instance}}",
        {{- end }}
        {{- if .istio }}
          istio="{{.istio}}",
        {{- end }}
        {{- if .istio_io_rev }}
          istio_io_rev="{{.istio_io_rev}}",
        {{- end }}
        {{- if .job }}
          job="{{.job}}",
        {{- end }}
        {{- if .namespace }}
          namespace="{{.namespace}}",
        {{- end }}
        {{- if .operator_istio_io_component }}
          operator_istio_io_component="{{.operator_istio_io_component}}",
        {{- end }}
        {{- if .pod }}
          pod="{{.pod}}",
        {{- end }}
        {{- if .pod_template_hash }}
          pod_template_hash="{{.pod_template_hash}}",
        {{- end }}
        {{- if .release }}
          release="{{.release}}",
        {{- end }}
        {{- if .request_protocol }}
          request_protocol="{{.request_protocol}}",
        {{- end }}
        {{- if .response_code }}
          response_code="{{.response_code}}",
        {{- end }}
        {{- if .response_flags }}
          response_flags="{{.response_flags}}",
        {{- end }}
        {{- if .service_istio_io_canonical_name }}
          service_istio_io_canonical_name="{{.service_istio_io_canonical_name}}",
        {{- end }}
        {{- if .service_istio_io_canonical_revision }}
          service_istio_io_canonical_revision="{{.service_istio_io_canonical_revision}}",
        {{- end }}
        {{- if .sidecar_istio_io_inject }}
          sidecar_istio_io_inject="{{.sidecar_istio_io_inject}}",
        {{- end }}
        {{- if .source_app }}
          source_app="{{.source_app}}",
        {{- end }}
        {{- if .source_canonical_revision }}
          source_canonical_revision="{{.source_canonical_revision}}",
        {{- end }}
        {{- if .source_canonical_service }}
          source_canonical_service="{{.source_canonical_service}}",
        {{- end }}
        {{- if .source_cluster }}
          source_cluster="{{.source_cluster}}",
        {{- end }}
        {{- if .source_principal }}
          source_principal="{{.source_principal}}",
        {{- end }}
        {{- if .source_version }}
          source_version="{{.source_version}}",
        {{- end }}
        {{- if .source_workload }}
          source_workload="{{.source_workload}}",
        {{- end }}
        {{- if .source_workload_namespace }}
          source_workload_namespace="{{.source_workload_namespace}}",
        {{- end }}
        {{"{{"}}- if .destination_workload {{"}}"}}
          destination_workload="{{"{{"}}.destination_workload{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
        {{"{{"}}- if .destination_workload_namespace {{"}}"}}
          destination_workload_namespace="{{"{{"}}.destination_workload_namespace{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)
  jqExpression: .data.result[0].value[1]