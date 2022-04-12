{{ define "experiment" -}}
{{- include "task.istio" . }}
{{- include "task.assess" . }}
{{ end }}