{{- define "env.kserve.service" }}

{{- $APP_NAME := .Release.Name }}
{{- $APP_NAMESPACE := .Release.Namespace }}
{{- if (and .Values.application .Values.application.metadata) }}
{{- $APP_NAME := .Values.application.metadata.name }}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
{{- end }}

apiVersion: v1
kind: Service
metadata:
  name: {{ $APP_NAME }}
spec:
  externalName: knative-local-gateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
{{- end }} {{- /* define "env.kserve.service" */}}