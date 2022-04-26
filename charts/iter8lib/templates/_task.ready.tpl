{{- define "task.ready.tn" }}
{{- /* Optional timeout from .Values.ready.timeout */ -}}
{{- if .Values.ready.timeout }}
timeout: {{ .Values.ready.timeout }}
{{- end }}
{{- /* Optional timeout from .Values.ready.namespace (non-Kubernetes experiment) or .Release.Namespace (Kubernetes experiment) */ -}}
{{ $namespace := "" }}
{{- if .Values.ready.namespace }}
{{ $namespace = .Values.ready.namespace }}
{{- end }}
{{- if .Release.Namespace }}
{{ $namespace = .Release.Namespace }}
{{- end }}
{{- /* if one of .Values.ready.namespace or .Release.Namespace */ -}}
{{- if $namespace }}
namespace: {{ $namespace }}
{{- end }}
{{- end }}

{{- define "task.ready" }}
{{- /* If user has specified a check for readiesss of a Kubernetes Service */ -}}
{{- if .Values.ready.service }}
- task: k8s-object-ready
  with:
    name: {{ .Values.ready.service }}
    version: v1
    resource: services
{{- include "task.ready.tn" . | indent 4 }}
{{- end }}
{{- /* If user has specified a check for readiesss of a Kubernetes Deployment */ -}}
{{- if .Values.ready.deploy }}
- task: k8s-object-ready
  with:
    name: {{ .Values.ready.deploy }}
    version: v1
    resource: deployments
    condition: Available
{{- include "task.ready.tn" . | indent 4 }}
{{- end }}
{{- end }}