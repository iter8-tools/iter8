{{- define "task.http" -}}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "http values object is nil" }}
{{- end }}
{{/* url must be defined or a url must be defined for each endpoint */}}
{{- if not .url }}
{{- if .endpoints }}
{{- range $endpointID, $endpoint := .endpoints }}
{{- if not $endpoint.url }}
{{- fail (print "endpoint \"" (print $endpointID "\" does not have a url parameter")) }}
{{- end }}
{{- end }}
{{- else }}
{{- fail "please set the url parameter or the endpoints parameter" }}
{{- end }}
{{- end }}
{{- /**************************/ -}}
{{- /* Perform the various setup steps before the main task */ -}}
{{- $vals := mustDeepCopy . }}
{{- if $vals.payloadURL }}
# task: download payload from payload URL
- run: |
    curl -o payload.dat {{ $vals.payloadURL }}
{{- $_ := set $vals "payloadFile" "payload.dat" }}
{{- end }}
{{- /**************************/ -}}
{{- /* Repeat above for each endpoint */ -}}
{{- range $endpointID, $endpoint := $vals.endpoints }}
{{- if $endpoint.payloadURL }}
{{- $payloadFile := print $endpointID "_payload.dat" }}
# task: download payload from payload URL for endpoint
- run: |
    curl -o {{ $payloadFile }} {{ $endpoint.payloadURL }}
{{- $_ := set $endpoint "payloadFile" $payloadFile }}
{{- end }}
{{- end }}
{{- /**************************/ -}}
{{- /* Warmup task if requested */ -}}
{{- if or .warmupNumRequests .warmupDuration }}
{{- $warmupVals := mustDeepCopy $vals }}
{{- if .warmupNumRequests }}
{{- $_ := set $warmupVals "numRequests" .warmupNumRequests }}
{{- else }}
{{- $_ := set $warmupVals "duration" .warmupDuration}}
{{- end }}
{{- /* replace warmup options a boolean */ -}}
{{- $_ := unset $warmupVals "warmupDuration" }}
{{- $_ := unset $warmupVals "warmupNumRequests" }}
{{- $_ := set $warmupVals "warmup" true }}
# task: generate warmup HTTP requests
# collect Iter8's built-in HTTP latency and error-related metrics
- task: http
  with:
{{ toYaml $warmupVals | indent 4 }}
{{- end }}
{{- /* warmup done */ -}}
{{- /**************************/ -}}
{{- /* Main task */ -}}
{{- /* remove warmup options if present */ -}}
{{- $_ := unset . "warmupDuration" }}
{{- $_ := unset . "warmupNumRequests" }}
# task: generate HTTP requests for app
# collect Iter8's built-in HTTP latency and error-related metrics
- task: http
  with:
{{ toYaml $vals | indent 4 }}
{{- end }}