{{- define "k.spec.secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-spec
  annotations:
    iter8.tools/revision: {{ .Release.Revision | quote }}
stringData:
  experiment.yaml: |
{{ include "experiment" . | indent 4 }}
{{- if .Values.providers }}
{{- range .Values.providers }}
  {{ print . ".metrics.yaml" }}: |
{{ include (print "metrics." .) $ | indent 4 }}
{{- end }}
{{- end }}
{{- end }}