{{- define "task.github" }}
{{- if .Values.github }}
# task: send a GitHub notification
- task: notify
  with:
    url: https://api.github.com/repos/{{ .Values.github.owner }}/{{ .Values.github.repo }}/dispatches
    method: POST
    headers:
      Accept: application/vnd.github.everest-preview+json
      Accept: "application/preview-github/json"
    payloadTemplateURL: "https://raw.githubusercontent.com/Alan-Cha/iter8/slack-notification/charts/iter8/templates/_payload-github.tpl"
    softFailure: {{ default true .Values.github.softFailure }}
{{- end }}
{{ end }}