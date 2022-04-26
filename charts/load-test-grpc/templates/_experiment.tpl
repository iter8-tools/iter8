{{- define "experiment" -}}
{{- include "task.ready" . -}}
{{- include "task.grpc" . -}}
{{- include "task.assess" . -}}
{{- end }}