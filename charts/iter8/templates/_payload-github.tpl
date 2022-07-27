{{- define "payload.github" }}
{
	"event_type": "iter8",
	"client_data": {{ .Values.JSONStringReport }}
}
{{- end }}