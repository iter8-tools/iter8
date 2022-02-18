{{ define "experiment" -}}
# task 1: generate gRPC requests for the given gRPC method
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:

    {{- if .Values.protoFile }}
    proto: {{ .Values.protoFile | toString }}
    {{- end }}

    {{- if .Values.protoURL }}
    protoURL: {{ .Values.protoURL | toString }}
    {{- end }}

    {{- if .Values.protosetFile }}
    protoset: {{ .Values.protosetFile | toString }}
    {{- end }}

    {{- if .Values.protosetURL }}
    protosetURL: {{ .Values.protosetURL | toString }}
    {{- end }}

    {{- if .Values.total }}
    total: {{ .Values.total | int }}
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

    {{- if .Values.duration }}
    duration: {{ .Values.duration | toString }}
    {{- end }}

    {{- if .Values.maxDuration }}
    max-duration: {{ .Values.maxDuration | toString }}
    {{- end }}

    {{- if .Values.streamInterval }}
    stream-interval: {{ .Values.streamInterval | toString }}
    {{- end }}

    {{- if .Values.streamCallDuration }}
    stream-call-duration: {{ .Values.streamCallDuration | toString }}
    {{- end }}

    {{- if .Values.streamCallCount }}
    stream-call-count: {{ .Values.streamCallCount | int }}
    {{- end }}

    {{- if .Values.connectTimeout }}
    connect-timeout: {{ .Values.connectTimeout | toString }}
    {{- end }}

    {{- if .Values.keepalive }}
    keepalive: {{ .Values.keepalive | toString }}
    {{- end }}

    {{- if .Values.data }}
    data:
{{ toYaml .Values.data | indent 6 }}
    {{- end }}

    {{- if .Values.dataFile }}
    data-file: {{ .Values.dataFile | toString }}
    {{- end }}

    {{- if .Values.dataURL }}
    dataURL: {{ .Values.dataURL | toString }}
    {{- end }}

    {{- if .Values.binaryDataFile }}
    binary-file: {{ .Values.binaryDataFile | toString }}
    {{- end }}

    {{- if .Values.binaryDataURL }}
    binaryDataURL: {{ .Values.binaryDataURL | toString }}
    {{- end }}

    {{- if .Values.metadata }}
    metadata:
{{ toYaml .Values.metadata | indent 6 }}
    {{- end }}

    {{- if .Values.metadataFile }}
    metadata-file: {{ .Values.metadataFile | toString }}
    {{- end }}

    {{- if .Values.metadataURL }}
    metadataURL: {{ .Values.metadataURL | toString }}
    {{- end }}

    {{- if .Values.reflectMetadata }}
    reflectMetadata:
{{ toYaml .Values.reflectMetadata | indent 6 }}
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
    - metric: {{ $key | toString }}
      upperLimit: {{ $value | float64 }}
    {{- end }}
{{- end }}
{{ end }}