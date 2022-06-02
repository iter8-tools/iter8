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
    - {{ .Values.versionInfo | toYaml | indent 6 | trim }}
{{- end }}