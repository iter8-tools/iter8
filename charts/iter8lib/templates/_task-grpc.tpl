{{- define "task.grpc" -}}
# task: generate gRPC requests for the given gRPC method
# collect Iter8's built-in gRPC latency and error-related metrics
- task: gen-load-and-collect-metrics-grpc
  with:
{{ toYaml .Values | indent 4 }}
{{- end }}    
