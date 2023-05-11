{{- define "task.custommetrics" }}
# task: collect custom metrics from providers (databases)
- task: custommetrics
  with:
{{ . | toYaml | indent 4 }}
{{- end }}