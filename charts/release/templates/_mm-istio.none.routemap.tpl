{{- define "env.mm-istio.none.routemap" }}

{{- $versions := include "normalize.versions.kserve-mm" . | mustFromJson }}

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
    {{- end }}

{{- end }} {{- /* define "env.mm-istio.none.routemap" */}}
