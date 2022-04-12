{{- define "task.istio" -}}
- task: collect-metrics-database
  with:
    versionInfo:
    - destination_workload: {{ required "A valid destination_workload value is required!" .Values.destination_workload | toString }}
      destination_workload_namespace: {{ required "A valid destination_workload_namespace value is required!" .Values.destination_workload_namespace | toString }}
      {{- if .Values.StartingTime }}
      StartingTime: {{ .Values.StartingTime | int }}
      {{- end }}

      # TODO: Should we make this more generic using the below?
      # {{ toYaml .Values.versionInfo | indent 8 }}
{{- end }}