{{- define "k.spec.secret-istio" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}-spec
stringData:
  experiment.yaml: |
    {{ include "experiment" .}}
  metrics.yaml: |
{{ include "istio.metrics" . }}
{{- end }}
