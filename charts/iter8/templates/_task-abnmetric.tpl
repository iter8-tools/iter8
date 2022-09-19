{{- define "task.abnmetrics" }}
# task: collect metrics written by Iter8 SDK
- task: abnmetrics
  with:
{{ . | toYaml | indent 4 }}
{{- end }}