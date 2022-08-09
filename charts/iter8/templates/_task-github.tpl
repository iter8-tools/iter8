{{- define "task.github" -}}
# task: send a GitHub notification
- task: notify
  with:
    url: https://api.github.com/repos/{{ .owner }}/{{ .repo }}/dispatches
    method: POST
    headers:
      Authorization: token {{ .token }}
      Accept: application/vnd.github+json
    payloadTemplateURL: "https://raw.githubusercontent.com/Alan-Cha/iter8/slack-notification/charts/iter8/templates/_payload-github.tpl"
    softFailure: {{ default true .softFailure }}
{{ end }}