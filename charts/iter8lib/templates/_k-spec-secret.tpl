{{- define "k.spec.secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-spec
  annotations:
    iter8.tools/revision: {{ .Release.Revision | toString }}
stringData:
  experiment.yaml: |
{{- include "experiment" . | indent 4 }}
{{- end }}
