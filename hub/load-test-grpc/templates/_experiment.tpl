{{ define "load-test-grpc.experiment" -}}
# task 1: generate gRPC requests for the given gRPC method
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:

    {{- if .Values.protoURL }}
    protoURL: "{{ .Values.protoURL}}"
    {{- end }}

    {{- if .Values.data }}
    data: {{ .Values.data | toString }}
    {{- end }}

    {{- ""}}
    versionInfo:
    - host: "{{ required "A valid host is required!" .Values.host }}"
      call: "{{ required "A valid call is required!" .Values.call }}"

{{- if .Values.SLOs }}
# task 2: validate service level objectives for app using
# the metrics collected in the above task
- task: assess-app-versions
  with:
    SLOs:
    {{- range $key, $value := .Values.SLOs }}
    {{- if or (regexMatch "error-rate" $key) (regexMatch "error-count" $key) }}
    - metric: "built-in/grpc-{{ $key }}"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "latency/max" $key) }}
    - metric: "built-in/grpc-latency/max"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "latency/stddev" $key) }}
    - metric: "built-in/grpc-latency/stddev"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "latency/mean" $key) }}
    - metric: "built-in/grpc-latency/mean"
      upperLimit: {{ $value }}
    {{- else if (regexMatch "^latency/p\\d+(?:\\.\\d)?$" $key) }}
    - metric: "built-in/grpc-{{ $key }}"
      upperLimit: {{ $value }}
    {{- else }}
    {{- fail "Invalid SLO metric specified" }}
    {{- end }}
    {{- end }}
{{- end }}
{{ end }}