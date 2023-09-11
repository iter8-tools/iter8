{{- define "k.role" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}
  annotations:
    iter8.tools/test: {{ .Release.Name }}
rules:
- apiGroups: [""]
  resourceNames: [{{ .Release.Name | quote }}]
  resources: ["secrets"]
  verbs: ["get", "update"]
{{- if .Values.ready }}
---
{{- $namespace := coalesce $.Values.ready.namespace $.Release.Namespace }}    
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}-ready
  {{- if $namespace }}
  namespace: {{ $namespace }}
  {{- end }} {{- /* if $namespace */}}
  annotations:
    iter8.tools/test: {{ .Release.Name }}
rules:
{{- $typesToCheck := omit .Values.ready "timeout" "namespace" }}
{{- range $type, $name := $typesToCheck }}
{{- $definition := get $.Values.resourceTypes $type }}
{{- if not $definition }}
{{- cat "no type definition for: " $type | fail }}
{{- else }}
- apiGroups: [ {{ get $definition "Group" | quote }} ]
  resourceNames: [ {{ $name | quote }} ]
  resources: [ {{ get $definition "Resource" | quote }} ]
  verbs: [ "get" ]
{{- end }} {{- /* if not $definition */}}
{{- end }} {{- /* range $type, $name */}}
{{- end }} {{- /* {{- if .Values.ready */}}
{{- end }} {{- /* {{- if .Values.ready */}}
