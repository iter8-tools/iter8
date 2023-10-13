{{- define "env.kserve.none.routemap" }}

{{- $versions := include "normalize.versions.kserve" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
{{- template "routemap.metadata" . }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: isvc
        name: {{ template "isvc.name" $v }}
        namespace: {{ template "isvc.namespace" $v }}
    {{- end }} {{- /* range $i, $v := $versions */}}

{{- end }} {{- /* define "env.kserve.none.routemap" */}}
