{{ define "experiment" -}}
{{- include "task.readiness" . }}
{{- include "task.http" . }}
{{- include "task.assess" . }}
{{ end }}