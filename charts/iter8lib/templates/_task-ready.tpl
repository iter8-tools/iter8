{{- define "task.ready.tn" }}
{{- /* Optional timeout from .Values.ready.timeout */ -}}
{{- if .Values.ready }}
{{- if .Values.ready.timeout }}
timeout: {{ .Values.ready.timeout }}
{{- end }}
{{- end }}
{{- /* Optional timeout from .Values.ready.namespace (non-Kubernetes experiment) or .Release.Namespace (Kubernetes experiment) */ -}}
{{ $namespace := "" }}
{{- if .Values.ready }}
{{- if .Values.ready.namespace }}
{{ $namespace = .Values.ready.namespace }}
{{- end }}
{{- end }}
{{- if .Release.Namespace }}
{{ $namespace = .Release.Namespace }}
{{- end }}
{{- /* if one of .Values.ready.namespace or .Release.Namespace */ -}}
{{- if $namespace }}
namespace: {{ $namespace }}
{{- end }}
{{ end }}

{{- define "task.ready" }}
{{- /* If user has specified a check for readiesss of a Kubernetes Service */ -}}
{{- if .Values.ready }}
{{- if .Values.ready.service }}
# task: determine if Kubernetes Service exists
- task: k8s-object-ready
  with:
    name: {{ .Values.ready.service | quote }}
    version: v1
    resource: services
{{- include "task.ready.tn" . | indent 4 }}
{{ end }}
{{- /* If user has specified a check for readiesss of a Kubernetes Deployment */ -}}
{{- if .Values.ready.deploy }}
# task: determine if Kubernetes Deployment is Available
- task: k8s-object-ready
  with:
    name: {{ .Values.ready.deploy | quote }}
    group: apps
    version: v1
    resource: deployments
    condition: Available
{{- include "task.ready.tn" . | indent 4 }}
{{ end }}
{{- end }}
{{- end }}