{{- define "env.deployment-istio.service" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}

apiVersion: v1
kind: Service
metadata:
  name: {{ $APP_NAME }}
  namespace: {{ $APP_NAMESPACE }}
spec:
  externalName: istio-ingressgateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
{{- end }} {{- /* define "env.deployment-istio.service" */}}