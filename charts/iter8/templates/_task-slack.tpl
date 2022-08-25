{{- define "task.slack" -}}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "slack notify values object is nil" }}
{{- end }}
{{- if not .url }}
{{- fail "please set a value for the url parameter" }}
{{- end }}
# task: send a Slack notification
- task: notify
  with:
    url: {{ .url }}
    method: POST
    payloadTemplateURL: {{ default "https://raw.githubusercontent.com/iter8-tools/iter8/v0.11.10/charts/iter8/templates/_payload-slack.tpl" .payloadTemplateURL }}
    softFailure: {{ default true .softFailure }}
{{ end }} 