{{- define "kserve.routemap-none" }}
{{- $versions := include "resolve.appVersions" . | mustFromJson }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.appName }}-routemap
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
        name: {{ default (printf "%s-%d" $.Values.appName $i) $v.name }}
        namespace: {{ default "modelmesh-serving" $v.namespace }}
    {{- end }}
immutable: true
{{- end }}
