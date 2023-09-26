{{- define "deployment.routemap-none" }}
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
      - gvrShort: svc
        name: {{ $v.name }}
        namespace: {{ $v.namespace }}
      - gvrShort: deploy
        name: {{ $v.name }}
        namespace: {{ $v.namespace }}
    {{- end }}
immutable: true
{{- end }}
