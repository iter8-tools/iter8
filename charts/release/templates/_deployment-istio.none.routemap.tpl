{{- define "env.deployment-istio.none.routemap" }}

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
        name: {{ template "svc.name" $v }}
        namespace: {{ template "svc.namespace" $v }}
      - gvrShort: deploy
        name: {{ template "deploy.name" $v }}
        namespace: {{ template "deploy.namespace" $v }}
    {{- end }}

{{- end }} {{- /* define "env.deployment-istio.none.routemap" */}}
