{{- define "k.serviceaccount" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Release.Name }}
{{- end }}
