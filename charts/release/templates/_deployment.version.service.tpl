{{- define "env.deployment.version.service" }}

{{- /* compute basic metadata */}}
{{- $metadata := include "application.version.metadata" . | mustFromJson }}

apiVersion: v1
kind: Service
{{- if .serviceSpecification }}
metadata:
{{- if .serviceSpecification.metadata }}
  {{ toYaml (merge .serviceSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .serviceSpecification.metadata */}}
spec:
  {{ toYaml .serviceSpecification.spec | nindent 2  | trim }}
{{- else }}
{{- if not .port }} {{- /* require .port */}}
{{- print "missing field: port required when serviceSpecification absent" | fail }}
{{- end }} {{- /* if not .port */}}
metadata:
  {{ toYaml $metadata | nindent 2 | trim }}
spec:
  selector:
    app: {{ .VERSION_NAME }}
  ports:
  - port: {{ .port }}
{{- end }} {{- /* if .serviceSpecification */}}
{{- end }} {{- /* define "env.deployment.version.service" */}}
