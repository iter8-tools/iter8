{{- define "task.github" -}}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "github notify values object is nil" }}
{{- end }}
{{- if not .owner }}
{{- fail "please set a value for the owner parameter" }}
{{- end }}
{{- if not .repo }}
{{- fail "please set a value for the repo parameter" }}
{{- end }}
{{- if not .token }}
{{- fail "please set a value for the token parameter" }}
{{- end }}
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