{{- define "task.grpc" -}}
{{/* Validate values */}}
{{- if not .Values.host }}
{{- fail "Please set a value for the host parameter." }}
{{- end }}
{{- if not .Values.call }}
{{- fail "Please set a value for the call parameter." }}
{{- end }}
{{/* Perform the various setup steps before the main task */}}
{{- $vals := mustDeepCopy .Values }}
{{- if .Values.protoURL }}
# task: download proto file from URL
- run: |
    curl -o ghz.proto {{ $vals.protoURL }}
{{- $pf := dict "proto" "ghz.proto" }}
{{- $vals = mustMerge $pf .Values }}
{{- end }}
{{- if .Values.dataURL }}
# task: download JSON data file from URL
- run: |
    curl -o data.json {{ $vals.dataURL }}
{{- $pf := dict "data-file" "data.json" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{- if .Values.binaryDataURL }}
# task: download binary data file from URL
- run: |
    curl -o data.bin {{ $vals.binaryDataURL }}
{{- $pf := dict "binary-file" "data.bin" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{- if .Values.metadataURL }}
# task: download metadata JSON file from URL
- run: |
    curl -o metadata.json {{ $vals.metadataURL }}
{{- $pf := dict "metadata-file" "metadata.json" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{/* Write the main task */}}
# task: generate gRPC requests for app
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:
{{ toYaml $vals | indent 4 }}
{{- end }}