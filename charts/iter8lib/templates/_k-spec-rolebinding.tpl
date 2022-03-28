{{- define "k.spec.rolebinding" -}}
{{- $name := printf "%v-%v" .Release.Name .Release.Revision -}}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ $name }}-spec-rolebinding
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: {{ $name }}-spec-role
  apiGroup: rbac.authorization.k8s.io
{{- end -}}
