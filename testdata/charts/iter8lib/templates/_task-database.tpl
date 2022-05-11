{{- define "task.database" -}}
- task: collect-metrics-database
  with:
    {{- if .Values.providers }}
    providers:
    {{- range .Values.providers }}
      - {{ . }}
    {{- end }}
    {{- end }}
    versionInfo:
    - endpoint: {{ required "A valid endpoint value is required!" .Values.endpoint | toString }}
      destination_workload: {{ required "A valid destination_workload value is required!" .Values.destination_workload | toString }}
      destination_workload_namespace: {{ required "A valid destination_workload_namespace value is required!" .Values.destination_workload_namespace | toString }}
      {{- if .Values.startingTime }}
      startingTime: {{ .Values.startingTime }}
      {{- end }}
{{- end }}