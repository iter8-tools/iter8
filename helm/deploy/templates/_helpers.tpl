{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "deploy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "deploy.labels.common" -}}
helm.sh/chart: {{ include "deploy.chart" . }}
app.kubernetes.io/managed-by: Iter8
app.kubernetes.io/name: {{ .Values.name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Labels for stable version
*/}}
{{- define "deploy.labels.stable" -}}
{{- include "deploy.labels.common" . }}
app.kubernetes.io/versionType: stable
{{- end }}

{{/*
Labels for candidate version
*/}}
{{- define "deploy.labels.candidate" -}}
{{- include "deploy.labels.common" . }}
app.kubernetes.io/versionType: candidate
{{- end }}
