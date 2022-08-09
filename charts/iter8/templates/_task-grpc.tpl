{{- define "task.grpc" }}
{{- /* Validate values */ -}}
{{- if not . }}
{{- fail "grpc values object is nil" }}
{{- end }}
{{- if not .host }}
{{- fail "please set a value for the host parameter" }}
{{- end }}
{{- if not .call }}
{{- fail "please set a value for the call parameter" }}
{{- end }}
{{- /* Perform the various setup steps before the main task */ -}}
{{- $vals := mustDeepCopy . }}
{{- if $vals.protoURL }}
# task: download proto file from URL
- run: |
    curl -o ghz.proto {{ $vals.protoURL }}
{{- $pf := dict "proto" "ghz.proto" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{- if $vals.dataURL }}
# task: download JSON data file from URL
- run: |
    curl -o data.json {{ $vals.dataURL }}
{{- $pf := dict "data-file" "data.json" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{- if $vals.binaryDataURL }}
# task: download binary data file from URL
- run: |
    curl -o data.bin {{ $vals.binaryDataURL }}
{{- $pf := dict "binary-file" "data.bin" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{- if $vals.metadataURL }}
# task: download metadata JSON file from URL
- run: |
    curl -o metadata.json {{ $vals.metadataURL }}
{{- $pf := dict "metadata-file" "metadata.json" }}
{{- $vals = mustMerge $pf $vals }}
{{- end }}
{{/* Write the main task */}}
# task: generate gRPC requests for app
# collect Iter8's built-in gRPC latency and error-related metrics
- task: grpc
  with:
{{ toYaml $vals | indent 4 }}
{{- end }}