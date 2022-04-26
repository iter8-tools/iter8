{{- define "k.spec.role" -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}-spec-role
  annotations:
    iter8.tools/revision: {{ .Release.Revision }}
rules:
- apiGroups: [""]
  resourceNames: ["{{ .Release.Name }}-spec"]
  resources: ["secrets"]
  verbs: ["get"]
{{- end }}
