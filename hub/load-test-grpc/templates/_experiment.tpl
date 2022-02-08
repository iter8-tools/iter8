{{ define "load-test-grpc.experiment" -}}
# task 1: generate gRPC requests for the given gRPC method
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:

    {{- if .Values.protoURL }}
    protoURL: {{ .Values.protoURL | toString }}
    {{- end }}

    {{- if .Values.connect-timeout }}
    connect-timeout: {{ .Values.connect-timeout | toString }}
    {{- end }}

    {{- if .Values.total }}
    total: {{ .Values.total | int }}
    {{- end }}

    {{- if .Values.max-duration }}
    max-duration: {{ .Values.max-duration | toString }}
    {{- end }}

    {{- if .Values.duration }}
    duration: {{ .Values.duration | toString }}
    {{- end }}

    {{- if .Values.rps }}
    rps: {{ .Values.rps | int }}
    {{- end }}

    {{- if .Values.concurrency }}
    concurrency: {{ .Values.concurrency | int }}
    {{- end }}

    {{- if .Values.connections }}
    connections: {{ .Values.connections | int }}
    {{- end }}

    {{- if .Values.data }}
    data:
{{ toYaml .Values.data | indent 6 }}
    {{- end }}

    {{- if .Values.dataURL }}
    dataURL: {{ .Values.dataURL | toString }}
    {{- end }}

    {{- if .Values.binaryDataURL }}
    binaryDataURL: {{ .Values.binaryDataURL | toString }}
    {{- end }}

    {{- if .Values.metadata }}
    metadata:
{{ toYaml .Values.metadata | indent 6 }}
    {{- end }}

    {{- if .Values.metadataURL }}
    metadataURL: {{ .Values.metadataURL | toString }}
    {{- end }}

    {{- ""}}
    versionInfo:
    - host: {{ required "A valid host is required!" .Values.host | toString }}
      call: {{ required "A valid call is required!" .Values.call | toString }}

{{- if .Values.SLOs }}
# task 2: validate service level objectives for app using
# the metrics collected in the above task
- task: assess-app-versions
  with:
    SLOs:
    {{- range $key, $value := .Values.SLOs }}
    {{- if or (regexMatch "error-rate" $key) (regexMatch "error-count" $key) }}
    - metric: "built-in/grpc-{{ $key }}"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "latency/max" $key) }}
    - metric: "built-in/grpc-latency/max"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "latency/stddev" $key) }}
    - metric: "built-in/grpc-latency/stddev"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "latency/mean" $key) }}
    - metric: "built-in/grpc-latency/mean"
      upperLimit: {{ $value | float64 }}
    {{- else if (regexMatch "^latency/p\\d+(?:\\.\\d)?$" $key) }}
    - metric: "built-in/grpc-{{ $key }}"
      upperLimit: {{ $value | float64 }}
    {{- else }}
    {{- fail "Invalid SLO metric specified" }}
    {{- end }}
    {{- end }}
{{- end }}
{{ end }}