{{- define "task.assess" -}}
{{- if . }}
# task: validate service level objectives for app using
# the metrics collected in an earlier task
- task: assess
  with:
{{- if .SLOs }}
    SLOs:
{{- if .SLOs.upper }}
      upper:
{{- range $m, $l := .SLOs.upper }}
      - metric: {{ $m }}
        limit: {{ $l }}
{{- end }}
{{- end }}
{{- if .SLOs.lower }}
      lower:
{{- range $m, $l := .SLOs.lower }}
      - metric: {{ $m }}
        limit: {{ $l }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{ end }}