{{- define "k.result.rolebinding" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}-result-rolebinding
  annotations:
    iter8.tools/revision: {{ .Release.Revision }}
subjects:
- kind: ServiceAccount
  name: default
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}-result-role
  apiGroup: rbac.authorization.k8s.io
{{- end }}
