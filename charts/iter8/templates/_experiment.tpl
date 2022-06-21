{{- define "experiment" -}}
{{- if not .Values.tasks }}
{{- fail ".Values.tasks is empty" }}
{{- end }}
spec:
  {{- range .Values.tasks }}
  {{- if eq "assess" . }}
  {{- include "task.assess" $.Values.assess -}}
  {{- else if eq "custommetrics" . }}
  {{- include "task.custommetrics" $.Values.custommetrics -}}
  {{- else if eq "grpc" . }}
  {{- include "task.grpc" $.Values.grpc -}}
  {{- else if eq "http" . }}
  {{- include "task.http" $.Values.http -}}
  {{- else if eq "ready" . }}
  {{- include "task.ready" $.Values.ready -}}
  {{- else }}
  {{- fail "task name must be one of assess, custommetrics, grpc, http, or ready" -}}
  {{- end }}
  {{- end }}
result:
  startTime:         {{ now | toJson }}
  numCompletedTasks: 0
  failure:           false
  iter8Version:      {{ .Values.majorMinor }}
{{- end }}