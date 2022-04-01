{{- define "k.spec.secret" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ $name }}-spec
stringData:
  experiment.yaml: |
{{ include "experiment" . | indent 4 }}
{{- end }}
