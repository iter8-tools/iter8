{{ define "experiment" -}}
{{- include "task.database" . }}
{{- include "task.assess" . }}
{{ end }}