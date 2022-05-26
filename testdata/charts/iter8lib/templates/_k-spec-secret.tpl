{{- define "k.spec.secret" -}}
{{ $globalContext := . }}
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
{{- if eq . "istio"}}
  istio.metrics.yaml: |
{{ include "metrics.istio" $globalContext | indent 4 }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}