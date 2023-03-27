git add {{- define "task.slack" -}}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "slack notify values object is nil" }}
{{- end }}
{{- if not .url }}
{{- fail "please set a value for the url parameter" }}
{{- end }}
# task: send a Slack notification
- task: notify
{{- if .if }}
  if: {{ .if | quote }}
{{- end }}
  with:
    url: {{ .url }}
    method: POST
    payloadTemplateURL: {{ default "https://raw.githubusercontent.com/iter8-tools/iter8/iter8-0.13.7/templates/notify/_payload-slack.tpl" .payloadTemplateURL }}
    softFailure: {{ default true .softFailure }}
{{ end }} 