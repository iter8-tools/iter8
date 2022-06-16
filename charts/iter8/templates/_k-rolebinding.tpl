{{- define "k.rolebinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}
  annotations:
    iter8.tools/group: {{ .Release.Name }}
subjects:
- kind: ServiceAccount
  name: {{ .Release.Name }}
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}
  apiGroup: rbac.authorization.k8s.io
{{- end }}
