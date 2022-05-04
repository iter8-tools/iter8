{{- define "task.ready.rbac" -}}
{{- if not .Values.iter8lib.disable.readinessrbac }}
{{- if .Values.ready }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .Release.Name }}-readiness-role
  annotations:
    iter8.tools/revision: {{ .Release.Revision | quote }}
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
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Release.Name }}-readiness-rolebinding
  annotations:
    iter8.tools/revision: {{ .Release.Revision | quote }}
subjects:
- kind: ServiceAccount
  name: default
  namespace: {{ .Release.Namespace }}
roleRef:
  kind: Role
  name: {{ .Release.Name }}-readiness-role
  apiGroup: rbac.authorization.k8s.io
{{- end }}
{{- end }}
{{- end }}