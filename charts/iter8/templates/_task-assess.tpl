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
{{- if .rewards }}
    rewards:
{{- if .rewards.max }}
      max:
{{- range $r, $val := .rewards.max }}
      - {{ $val }}
{{- end }}
{{- end }}
{{- if .rewards.min }}
      min:
{{- range $r, $val := .rewards.min }}
      - {{ $val }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}