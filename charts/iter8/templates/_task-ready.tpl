{{- define "task.ready" }}
{{- if .Values.ready }}
{{- $namespace := coalesce .Values.ready.namespace .Release.Namespace }}
{{- if .Values.ready.service }}
# task: determine if Kubernetes Service exists
- task: ready
  with:
    name: {{ .Values.ready.service | quote }}
    version: v1
    resource: services
{{- if $namespace }}
    namespace: {{ $namespace }}
{{- end }}
{{- if .Values.ready.timeout }}
    timeout: {{ .Values.ready.timeout }}
{{- end }}
{{- end }}
{{- if .Values.ready.deploy }}
# task: determine if Kubernetes Deployment exists and is Available
- task: ready
  with:
    name: {{ .Values.ready.deploy | quote }}
    group: apps
    version: v1
    resource: deployments
    condition: Available
{{- if $namespace }}
    namespace: {{ $namespace }}
{{- end }}
{{- if .Values.ready.timeout }}
    timeout: {{ .Values.ready.timeout }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}