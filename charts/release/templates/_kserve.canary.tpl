{{- define "env.kserve.canary.routemap" }}

{{- $APP_NAME := .Release.Name }}
{{- $APP_NAMESPACE := .Release.Namespace }}
{{- if (and .Values.application .Values.application.metadata) }}
{{- $APP_NAME := .Values.application.metadata.name }}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
{{- end }}
{{- $versions := include "normalize.versions" . | mustFromJson }}

{{- end }} {{- /* define "env.kserve.canary.routemap" */}}