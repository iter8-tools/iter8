{{- define "env.mm-istio.service" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}

apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: {{ $APP_NAME }}
  namespace: {{ $APP_NAMESPACE }}
spec:
  hosts:
  - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
  - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
  - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
  location: MESH_INTERNAL
  ports:
  - number: {{ template "mm.servicePort" . }}
    name: http
    protocol: HTTP
  resolution: DNS
  workloadSelector:
    labels:
      modelmesh-service: modelmesh-serving
{{- end }} {{- /* define "env.mm-istio.service" */}}