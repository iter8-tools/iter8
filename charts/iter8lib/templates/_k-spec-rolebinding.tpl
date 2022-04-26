{{- define "k.spec.rolebinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}-spec-rolebinding
  annotations:
    iter8.tools/revision: {{ .Release.Revision | toString }}
subjects:
- kind: ServiceAccount
  name: default
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}-spec-role
  apiGroup: rbac.authorization.k8s.io
{{- end }}
