{{- define "env.deployment-gtw.service" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}

apiVersion: v1
kind: Service
metadata:
  name: {{ $APP_NAME }}
  namespace: {{ $APP_NAMESPACE }}
spec:
  selector:
    app: {{ $APP_NAME }}
  ports:
  - port: 80
{{- end }} {{- /* define "env.deployment-gtw.service" */}}