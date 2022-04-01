{{- define "task.http" -}}
# task: generate HTTP requests for application URL
# collect Iter8's built-in HTTP latency and error-related metrics
- task: gen-load-and-collect-metrics-http
  with:

    {{- if .Values.numQueries }}
    numRequests: {{ .Values.numQueries | int }}
    {{- end }}

    {{- if .Values.duration }}
    duration: {{ .Values.duration | toString }}
    {{- end }}

    {{- if .Values.qps }}
    qps: {{ .Values.qps | float64 }}
    {{- end }}

    {{- if .Values.connections }}
    connections: {{ .Values.connections | int }}
    {{- end }}

    {{- if .Values.payloadStr }}
    payloadStr: {{ .Values.payloadStr | quote }}
    {{- end }}

    {{- if .Values.contentType }}
    contentType: {{ .Values.contentType | quote }}
    {{- end }}

    {{- if .Values.errorsAbove }}
    errorRanges:
    - lower: {{ .Values.errorsAbove | int }}
    {{- end }}

    {{- $percentiles := list }}
    {{- range $key, $value := .Values.SLOs }}
    {{- if (regexMatch "http/latency-p\\d+(?:\\.\\d)?$" $key) }}
    {{- $percentiles = append $percentiles (trimPrefix "http/latency-p" $key | float64 ) }}
    {{- end }}
    {{- end }}
    {{- if $percentiles }}
    percentiles:
{{ toYaml ($percentiles | uniq) | indent 4 }}
    {{- end }}

    {{- ""}}
    versionInfo:
    - url: {{ required "A valid url value is required!" .Values.url | toString }}
    {{- if .Values.headers }}
      headers:
{{ toYaml .Values.headers | indent 8 }}
    {{- end }}
{{- end }}    

