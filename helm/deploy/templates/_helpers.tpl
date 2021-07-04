{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "deploy.common.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels. Defined as a template instead of values in Values.yaml.
*/}}
{{- define "deploy.common.labels" -}}
helm.sh/chart: {{ include "deploy.common.chart" . }}
app.kubernetes.io/managed-by: Iter8
app.kubernetes.io/name: {{ .Values.common.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
