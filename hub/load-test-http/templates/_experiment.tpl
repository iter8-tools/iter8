{{ define "load-test.experiment" -}}
# task 1: generate HTTP requests for application URL
# collect Iter8's built-in HTTP latency and error-related metrics
- task: gen-load-and-collect-metrics-http
  with:

    {{- if .Values.numQueries }}
    numQueries: {{ .Values.numQueries}}
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

    {{- if .Values.errorAbove }}
    errorRanges:
    - lower: {{ .Values.errorAbove }}
    {{- end }}

    {{- $percentiles := list }}
    {{- range $key, $value := .Values.SLOs }}
    {{- if (regexMatch "^latency-p\\d+(?:\\.\\d)?$" $key) }}
    {{- $percentiles = append $percentiles (trimPrefix "latency-p" $key | float64 ) }}
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
    - metric: "built-in/http-{{ $key }}"
      upperLimit: {{ $value }}
    {{- end }}
{{- end }}
{{ end }}