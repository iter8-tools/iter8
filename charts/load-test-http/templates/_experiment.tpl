{{- define "experiment" -}}
{{- include "task.ready" . -}}
{{- include "task.http" . -}}
{{- include "task.assess" . -}}
{{- end }}