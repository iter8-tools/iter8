{{- define "create.weight-config" }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .name }}-weight-config
  labels:
    iter8.tools/watch: "true"
{{- if .weight }}
  annotations:
    iter8.tools/weight: "{{ .weight }}"
{{- end }}
{{- end }}
