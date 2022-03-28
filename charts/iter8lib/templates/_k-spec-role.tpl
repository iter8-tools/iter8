{{- define "k.spec.role" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ $name }}-spec-role
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["{{ $name }}-spec"]
  verbs: ["get"]
{{- end -}}
