{{ define "load-test.experiment" -}}
# task 1: generate HTTP requests for application URL
# collect Iter8's built-in HTTP latency and error-related metrics
- task: gen-load-and-collect-metrics-http
  with:

    {{- if .Values.numQueries }}
    numRequests: {{ .Values.numQueries}}
    {{- end }}

    {{- if .Values.duration }}
    duration: {{ .Values.duration}}
    {{- end }}

    {{- if .Values.qps }}
    qps: {{ .Values.qps}}
    {{- end }}

    {{- if .Values.connections }}
    connections: {{ .Values.connections}}
    {{- end }}

    {{- if .Values.payloadStr }}
    payloadStr: {{ .Values.payloadStr}}
    {{- end }}

    {{- if .Values.payloadURL }}
    payloadURL: {{ .Values.payloadURL}}
    {{- end }}

    {{- if .Values.contentType }}
    contentType: {{ .Values.contentType}}
    {{- end }}

    {{- if .Values.errorRanges }}
    errorRanges:
{{ toYaml .Values.errorRanges | indent 4 }}
    {{- end }}

    {{- $percentiles := list }}
    {{- if .Values.percentiles }}
    {{- $percentiles = concat $percentiles .Values.percentiles }}
    {{- end }}
    {{- range $key, $value := .Values.SLOs }}
    {{- if (regexMatch "^p\\d+(?:\\.\\d)?$" $key) }}
    {{- $percentiles = append $percentiles (trimPrefix "p" $key | float64 ) }}    
    {{- end }}
    {{- end }}
    {{- if $percentiles }}
    percentiles: 
{{ toYaml ($percentiles | uniq) | indent 4 }}
    {{- end }}

    {{- ""}}
    versionInfo:
    - url: {{ required "A valid url value is required!" .Values.url }}
    {{- if .Values.headers }}
    headers:
{{ toYaml .Values.headers | indent 6 }}
    {{- end }}

{{- if .Values.SLOs }}
# task 2: validate service level objectives for app using
# the metrics collected in the above task
- task: assess-app-versions
  with:
    SLOs:
    {{- range $key, $value := .Values.SLOs }}
    {{- if or (regexMatch "error-rate" $key) (regexMatch "error-count" $key) }}
    - metric: "built-in/http-{{ $key }}"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "mean-latency" $key) }}
    - metric: "built-in/http-latency-mean"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "^p\\d+(?:\\.\\d)?$" $key) }}
    - metric: "built-in/http-latency-{{ $key }}"
      upperLimit: {{ $value }}
    {{- else }}
    {{- fail "Invalid SLO metric specified" }}
    {{- end }}
    {{- end }}
{{- end }}
{{ end }}