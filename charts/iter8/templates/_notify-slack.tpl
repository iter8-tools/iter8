{{- define "task.slack" }}
{{- if .Values.slack }}
# task: send a Slack notification
- task: notify
  with:
    url: {{ .Values.slack.hook }}
    method: POST
    payloadTemplateURL: {{ default "https://github.com/iter8-tools/iter8/master/blob/slack-template.tpl" .Values.slack. payloadTemplateURL }}
    softFailure: {{ default true .Values.slack.softFailure }} 
{{- end }}
{{- end }}