{{- define "k.serviceaccount" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Release.Name }}-iter8-sa
{{- end }}
