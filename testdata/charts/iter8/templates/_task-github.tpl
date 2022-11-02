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
{{- if .if }}
  if: {{ .if | quote }}
{{- end }}
  with:
    url: https://api.github.com/repos/{{ .owner }}/{{ .repo }}/dispatches
    method: POST
    headers:
      Authorization: token {{ .token }}
      Accept: application/vnd.github+json
    payloadTemplateURL: {{ default "https://raw.githubusercontent.com/iter8-tools/hub/iter8-0.12.1/charts/iter8/templates/_payload-github.tpl" .payloadTemplateURL }}
    softFailure: {{ default true .softFailure }}
{{ end }}