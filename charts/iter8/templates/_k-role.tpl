{{- define "k.role" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
rules:
- apiGroups: [""]
  resourceNames: [{{ .Release.Name | quote }}]
  resources: ["secrets"]
  verbs: ["get", "update"]
{{- if .Values.ready }}
---
{{- template "task.ready.namespace" . }}
{{- if $namespace }}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}-ready
  namespace: {{ $namespace }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
rules:
{{- if .Values.ready.service }}
- apiGroups: [""]
  resourceNames: [{{ .Values.ready.service | quote }}]
  resources: ["services"]
  verbs: ["get"]
{{- end }}
{{- if .Values.ready.deploy }}
- apiGroups: ["apps"]
  resourceNames: [{{ .Values.ready.deploy | quote }}]
  resources: ["deployments"]
  verbs: ["get"]
{{- end }}
{{- end }}
{{- end }}
{{- end }}
