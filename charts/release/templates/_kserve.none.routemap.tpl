{{- define "env.kserve.none.routemap" }}

{{- $versions := include "normalize.versions" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
{{ template "routemap.metadata" . }}
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
