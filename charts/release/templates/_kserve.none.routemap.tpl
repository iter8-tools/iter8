{{- define "env.kserve.none.routemap" }}
{{- $APP_NAME := .Release.Name }}
{{- $APP_NAMESPACE := .Release.Namespace }}
{{- if (and .Values.application .Values.application.metadata) }}
{{- $APP_NAME := .Values.application.metadata.name }}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
{{- end }}
{{- $versions := include "normalize.versions" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $APP_NAME }}-routemap
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: {{ .Values.iter8Version }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: isvc
        name: {{ $v.VERSION_NAME }}
        namespace: {{ $v.VERSION_NAMESPACE }}
    {{- end }} {{- /* range $i, $v := .Values.application.versions */}}

{{- end }} {{- /* define "env.kserve.none.routemap" */}}
