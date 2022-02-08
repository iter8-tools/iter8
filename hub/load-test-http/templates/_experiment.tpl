{{ define "load-test-http.experiment" -}}
# task 1: generate HTTP requests for application URL
# collect Iter8's built-in HTTP latency and error-related metrics
- task: gen-load-and-collect-metrics-http
  with:

    {{- if .Values.numQueries }}
    numRequests: {{ .Values.numQueries | int }}
    {{- end }}

    {{- if .Values.duration }}
    duration: {{ .Values.duration | toString | quote }}
    {{- end }}

    {{- if .Values.qps }}
    qps: {{ .Values.qps | float64 }}
    {{- end }}

    {{- if .Values.connections }}
    connections: {{ .Values.connections | int }}
    {{- end }}

    {{- if .Values.payloadStr }}
    payloadStr: "{{ .Values.payloadStr | toString }}"
    {{- end }}

    {{- if .Values.payloadURL }}
    payloadURL: "{{ .Values.payloadURL | toString }}"
    {{- end }}

    {{- if .Values.contentType }}
    contentType: "{{ .Values.contentType | toString }}"
    {{- end }}

    {{- if .Values.errorsAbove }}
    errorRanges:
    - lower: {{ .Values.errorsAbove | int }}
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
    - url: {{ required "A valid url value is required!" .Values.url | toString | quote }}
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
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "latency-mean" $key) }}
    - metric: "built-in/http-latency-mean"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "latency-stddev" $key) }}
    - metric: "built-in/http-latency-stddev"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "^latency-p\\d+(?:\\.\\d)?$" $key) }}
    - metric: "built-in/http-{{ $key }}"
      upperLimit: {{ $value | float64 }}
    {{- else }}
    {{- fail "Invalid SLO metric specified" }}
    {{- end }}
    {{- end }}
{{- end }}
{{ end }}