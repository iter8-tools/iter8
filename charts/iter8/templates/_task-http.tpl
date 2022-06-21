{{- define "task.http" -}}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "http values object is nil" }}
{{- end }}
{{- if not .url }}
  {{- fail "please specify the url parameter" }}
{{- end }}
{{- /* Perform the various setup steps before the main task */ -}}
{{- $vals := mustDeepCopy . }}
{{- if $vals.payloadURL }}
# task: download payload from payload URL
- run: |
    curl -o payload.dat {{ $vals.payloadURL }}
{{- $pf := dict "payloadFile" "payload.dat" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{/* Write the main task */}}
# task: generate HTTP requests for app
# collect Iter8's built-in HTTP latency and error-related metrics
- task: http
  with:
{{ toYaml $vals | indent 4 }}
{{- end }}
