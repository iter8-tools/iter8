{{- define "k.rolebinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
subjects:
- kind: ServiceAccount
  name: {{ .Release.Name }}-iter8-sa
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}
  apiGroup: rbac.authorization.k8s.io
{{- if .Values.ready }}
---
{{- $namespace := coalesce .Values.ready.namespace .Release.Namespace }}
{{- if $namespace }}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}-ready
  namespace: {{ $namespace }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
subjects:
- kind: ServiceAccount
  name: {{ .Release.Name }}-iter8-sa
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}-ready
  apiGroup: rbac.authorization.k8s.io
{{- end }}
{{- end }}
{{- end }}
