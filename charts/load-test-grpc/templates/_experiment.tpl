{{ define "experiment" -}}
{{- include "task.readiness" . }}
{{- include "task.grpc" . -}}
{{- include "task.assess" . -}}
{{ end }}