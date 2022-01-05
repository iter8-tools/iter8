{{ define "load-test.experiment" -}}
# task 1: generate HTTP requests for application URL
# collect Iter8's built-in latency and error-related metrics
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: {{ .Values.url }}
# task 2: validate service level objectives for app using
# the metrics collected in the above task
- task: assess-app-versions
  with:
  {{- if .Values.SLOs }}
    SLOs:
{{ toYaml .Values.SLOs | indent 4 }}
  {{- end }}  
{{ end }}