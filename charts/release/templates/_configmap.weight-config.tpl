{{- define "configmap.weight-config" }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .VERSION_NAME }}-weight-config
  labels:
    iter8.tools/watch: "true"
{{- if .weight }}
  annotations:
    iter8.tools/weight: "{{ .weight }}"
{{- end }} {{- /* if .weight */}}
{{- end }} {{- /* define "configmap.weight-config" */}}
