{{- define "payload.slack" }}
{
 	"text": {{ .Values.JSONStringReport }}
}
{{- end }}