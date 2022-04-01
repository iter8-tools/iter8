{{- define "task.grpc" -}}
# task: generate gRPC requests for the given gRPC method
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:

    {{- if .Values.protoFile }}
    proto: {{ .Values.protoFile | quote }}
    {{- end }}

    {{- if .Values.protosetFile }}
    protoset: {{ .Values.protosetFile | quote }}
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
    duration: {{ .Values.duration | quote }}
    {{- end }}

    {{- if .Values.maxDuration }}
    max-duration: {{ .Values.maxDuration | quote }}
    {{- end }}

    {{- if .Values.streamInterval }}
    stream-interval: {{ .Values.streamInterval | quote }}
    {{- end }}

    {{- if .Values.streamCallDuration }}
    stream-call-duration: {{ .Values.streamCallDuration | quote }}
    {{- end }}

    {{- if .Values.streamCallCount }}
    stream-call-count: {{ .Values.streamCallCount | int }}
    {{- end }}

    {{- if .Values.connectTimeout }}
    connect-timeout: {{ .Values.connectTimeout | quote }}
    {{- end }}

    {{- if .Values.keepalive }}
    keepalive: {{ .Values.keepalive | quote }}
    {{- end }}

    {{- if .Values.data }}
    data:
{{ toYaml .Values.data | indent 6 }}
    {{- end }}

    {{- if .Values.dataFile }}
    data-file: {{ .Values.dataFile | quote }}
    {{- end }}

    {{- if .Values.binaryDataFile }}
    binary-file: {{ .Values.binaryDataFile | quote }}
    {{- end }}

    {{- if .Values.metadata }}
    metadata:
{{ toYaml .Values.metadata | indent 6 }}
    {{- end }}

    {{- if .Values.metadataFile }}
    metadata-file: {{ .Values.metadataFile | quote }}
    {{- end }}

    {{- if .Values.reflectMetadata }}
    reflectMetadata:
{{ toYaml .Values.reflectMetadata | indent 6 }}
    {{- end }}

    {{- ""}}
    versionInfo:
    - host: {{ required "A valid host is required!" .Values.host | quote }}
      call: {{ required "A valid call is required!" .Values.call | quote }}
{{- end }}    

