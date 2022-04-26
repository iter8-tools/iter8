{{- define "k.result.role" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}-result-role
  annotations:
    iter8.tools/revision: {{ .Release.Revision }}
rules:
- apiGroups: [""]
  resourceNames: ["{{ .Release.Name }}-result"]
  resources: ["secrets"]
  verbs: ["get", "update"]
{{- end }}
