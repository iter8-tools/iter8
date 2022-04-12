{{- define "task.http" -}}
{{/* Validate values */}}
{{- if not .Values.url }}
  {{- fail "Please set a value for the url parameter." }}
{{- end }}
{{/* Perform the various setup steps before the main task */}}
{{- $vals := mustDeepCopy .Values }}
{{- if $vals.percentiles }}
  {{- $percentiles := list }}
  {{- range $pval := $vals.percentiles }}
    {{- $percentiles = append $percentiles ( $pval | float64 ) }}
  {{- end }}
  {{ $_ = set $vals "percentiles" $percentiles }}
{{- end }}
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
- task: gen-load-and-collect-metrics-http
  with:
{{ toYaml $vals | indent 4 }}
{{- end }}
