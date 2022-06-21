{{- define "k.secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
stringData:
  experiment.yaml: |
{{ include "experiment" . | indent 4 }}
{{- end }}