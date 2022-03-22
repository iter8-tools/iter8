# endpoint where the monitoring instance is available
# https://cloud.ibm.com/docs/monitoring?topic=monitoring-endpoints#endpoints_sysdig
url: {{ .MonitoringEndpoint }}/prometheus/api/v1/query # e.g. https://ca-tor.monitoring.cloud.ibm.com
headers:
  # IAM token
  # to get the token, run: ibmcloud iam oauth-tokens | grep IAM | cut -d \: -f 2 | sed 's/^ *//'
  Authorization: Bearer {{ .IAMToken }}
  # GUID of the IBM Cloud Monitoring instance
  # to get the GUID, run: ibmcloud resource service-instance <NAME> --output json | jq -r '.[].guid'
  # https://cloud.ibm.com/docs/monitoring?topic=monitoring-mon-curl
  IBMInstanceID: {{ .GUID }}
provider: IBM Cloud Code Engine Sysdig
method: GET
# Inputs for the template:
#   ibm_codeengine_application_name string
#   ibm_codeengine_gateway_instance string
#   ibm_codeengine_namespace        string
#   ibm_codeengine_project_name     string
#   ibm_codeengine_revision_name    string
#   ibm_codeengine_status           string
#   ibm_ctype                       string
#   ibm_location                    string
#   ibm_scope                       string
#   ibm_service_instance            string
#   ibm_service_name                string
#
# Inputs for the metrics (output of template):
#   ibm_codeengine_revision_name string
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
      sum(last_over_time(ibm_codeengine_application_requests_total{
        {{- if .ibm_codeengine_application_name }}
          ibm_codeengine_application_name="{{.ibm_codeengine_application_name}}",
        {{- end }}
        {{- if .ibm_codeengine_gateway_instance }}
          ibm_codeengine_gateway_instance="{{.ibm_codeengine_gateway_instance}}",
        {{- end }}
        {{- if .ibm_codeengine_namespace }}
          ibm_codeengine_namespace="{{.ibm_codeengine_namespace}}",
        {{- end }}
        {{- if .ibm_codeengine_project_name }}
          ibm_codeengine_project_name="{{.ibm_codeengine_project_name}}",
        {{- end }}
        {{- if .ibm_codeengine_status }}
          ibm_codeengine_status="{{.ibm_codeengine_status}}",
        {{- end }}
        {{- if .ibm_ctype }}
          ibm_ctype="{{.ibm_ctype}}",
        {{- end }}
        {{- if .ibm_location }}
          ibm_location="{{.ibm_location}}",
        {{- end }}
        {{- if .ibm_scope }}
          ibm_scope="{{.ibm_scope}}",
        {{- end }}
        {{- if .ibm_service_instance }}
          ibm_service_instance="{{.ibm_service_instance}}",
        {{- end }}
        {{- if .ibm_service_name }}
          ibm_service_name="{{.ibm_service_name}}",
        {{- end }}
        {{"{{"}}- if .ibm_codeengine_revision_name {{"}}"}}
          ibm_codeengine_revision_name="{{"{{"}}.ibm_codeengine_revision_name{{"}}"}}",
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
      sum(last_over_time(ibm_codeengine_application_requests_total{
        ibm_codeengine_status!="200",
        {{- if .ibm_codeengine_application_name }}
          ibm_codeengine_application_name="{{.ibm_codeengine_application_name}}",
        {{- end }}
        {{- if .ibm_codeengine_gateway_instance }}
          ibm_codeengine_gateway_instance="{{.ibm_codeengine_gateway_instance}}",
        {{- end }}
        {{- if .ibm_codeengine_namespace }}
          ibm_codeengine_namespace="{{.ibm_codeengine_namespace}}",
        {{- end }}
        {{- if .ibm_codeengine_project_name }}
          ibm_codeengine_project_name="{{.ibm_codeengine_project_name}}",
        {{- end }}
        {{- if .ibm_codeengine_status }}
          ibm_codeengine_status="{{.ibm_codeengine_status}}",
        {{- end }}
        {{- if .ibm_ctype }}
          ibm_ctype="{{.ibm_ctype}}",
        {{- end }}
        {{- if .ibm_location }}
          ibm_location="{{.ibm_location}}",
        {{- end }}
        {{- if .ibm_scope }}
          ibm_scope="{{.ibm_scope}}",
        {{- end }}
        {{- if .ibm_service_instance }}
          ibm_service_instance="{{.ibm_service_instance}}",
        {{- end }}
        {{- if .ibm_service_name }}
          ibm_service_name="{{.ibm_service_name}}",
        {{- end }}
        {{"{{"}}- if .ibm_codeengine_revision_name {{"}}"}}
          ibm_codeengine_revision_name="{{"{{"}}.ibm_codeengine_revision_name{{"}}"}}",
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
      sum(last_over_time(ibm_codeengine_application_requests_total{
        ibm_codeengine_status!="200",
        {{- if .ibm_codeengine_application_name }}
          ibm_codeengine_application_name="{{.ibm_codeengine_application_name}}",
        {{- end }}
        {{- if .ibm_codeengine_gateway_instance }}
          ibm_codeengine_gateway_instance="{{.ibm_codeengine_gateway_instance}}",
        {{- end }}
        {{- if .ibm_codeengine_namespace }}
          ibm_codeengine_namespace="{{.ibm_codeengine_namespace}}",
        {{- end }}
        {{- if .ibm_codeengine_project_name }}
          ibm_codeengine_project_name="{{.ibm_codeengine_project_name}}",
        {{- end }}
        {{- if .ibm_codeengine_status }}
          ibm_codeengine_status="{{.ibm_codeengine_status}}",
        {{- end }}
        {{- if .ibm_ctype }}
          ibm_ctype="{{.ibm_ctype}}",
        {{- end }}
        {{- if .ibm_location }}
          ibm_location="{{.ibm_location}}",
        {{- end }}
        {{- if .ibm_scope }}
          ibm_scope="{{.ibm_scope}}",
        {{- end }}
        {{- if .ibm_service_instance }}
          ibm_service_instance="{{.ibm_service_instance}}",
        {{- end }}
        {{- if .ibm_service_name }}
          ibm_service_name="{{.ibm_service_name}}",
        {{- end }}
        {{"{{"}}- if .ibm_codeengine_revision_name {{"}}"}}
          ibm_codeengine_revision_name="{{"{{"}}.ibm_codeengine_revision_name{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0)/sum(last_over_time(ibm_codeengine_application_requests_total{
        {{- if .ibm_codeengine_application_name }}
          ibm_codeengine_application_name="{{.ibm_codeengine_application_name}}",
        {{- end }}
        {{- if .ibm_codeengine_gateway_instance }}
          ibm_codeengine_gateway_instance="{{.ibm_codeengine_gateway_instance}}",
        {{- end }}
        {{- if .ibm_codeengine_namespace }}
          ibm_codeengine_namespace="{{.ibm_codeengine_namespace}}",
        {{- end }}
        {{- if .ibm_codeengine_project_name }}
          ibm_codeengine_project_name="{{.ibm_codeengine_project_name}}",
        {{- end }}
        {{- if .ibm_codeengine_status }}
          ibm_codeengine_status="{{.ibm_codeengine_status}}",
        {{- end }}
        {{- if .ibm_ctype }}
          ibm_ctype="{{.ibm_ctype}}",
        {{- end }}
        {{- if .ibm_location }}
          ibm_location="{{.ibm_location}}",
        {{- end }}
        {{- if .ibm_scope }}
          ibm_scope="{{.ibm_scope}}",
        {{- end }}
        {{- if .ibm_service_instance }}
          ibm_service_instance="{{.ibm_service_instance}}",
        {{- end }}
        {{- if .ibm_service_name }}
          ibm_service_name="{{.ibm_service_name}}",
        {{- end }}
        {{"{{"}}- if .ibm_codeengine_revision_name {{"}}"}}
          ibm_codeengine_revision_name="{{"{{"}}.ibm_codeengine_revision_name{{"}}"}}",
        {{"{{"}}- end {{"}}"}}
      }[{{"{{"}}.ElapsedTime{{"}}"}}s])) or on() vector(0) 
  jqExpression: .data.result.[0].value.[1]