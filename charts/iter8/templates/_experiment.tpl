{{- define "experiment" -}}
{{- if not .Values.tasks }}
{{- fail ".Values.tasks is empty" }}
{{- end }}
metadata:
  name: {{ .Release.Name }}
  namespace: {{ .Release.Namespace }}
spec:
  {{- range .Values.tasks }}
  {{- if eq "grpc" . }}
  {{- include "task.grpc" $.Values.grpc -}}
  {{- else if eq "http" . }}
  {{- include "task.http" $.Values.http -}}
  {{- else if eq "ready" . }}
  {{- include "task.ready" $ -}}
  {{- else if eq "slack" . }}
  {{- include "task.slack" $.Values.slack -}}
  {{- else if eq "github" . }}
  {{- include "task.github" $.Values.github -}}
  {{- else }}
  {{- fail "task name must be one of grpc, http, ready, github, or slack" -}}
  {{- end }}
  {{- end }}
result:
  startTime:         {{ now | toJson }}
  numCompletedTasks: 0
  failure:           false
  iter8Version:      {{ .Values.majorMinor }}
{{- end }}