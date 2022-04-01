{{- define "task.assess" -}}
{{- if .Values.SLOs }}
# task: validate service level objectives for app using
# the metrics collected in an earlier task
- task: assess-app-versions
  with:
    SLOs:
    {{- range $key, $value := .Values.SLOs }}
    - metric: {{ $key | toString }}
      upperLimit: {{ $value | float64 }}
    {{- end }}
{{- end }}
{{- end }}

