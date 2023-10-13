{{- define "env.deployment-istio.canary.routemap" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
{{- template "routemap.metadata" . }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: svc
        name: {{ template "svc.name" $v}}
        namespace: {{ template "svc.namespace" $v}}
      - gvrShort: deploy
        name: {{ template "deploy.name" $v}}
        namespace: {{ template "deploy.namespace" $v}}
    {{- end }}

TBD: _deployment-istio.canary.routemap.tpl

{{- end }} {{- /* define "env.deployment-istio.canary.routemap" */}}
