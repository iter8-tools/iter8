{{- define "task.slack" -}}
# task: send a Slack notification
- task: notify
  with:
    url: {{ .url }}
    method: POST
    payloadTemplateURL: {{ default "https://github.com/iter8-tools/iter8/master/blob/slack-template.tpl" .payloadTemplateURL }}
    softFailure: {{ default true .softFailure }}
{{ end }}