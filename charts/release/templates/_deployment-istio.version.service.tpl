{{- define "env.deployment-istio.version.service" }}

{{- /* compute labels */}}
{{- $labels := include "application.version.labels" . | mustFromJson }}

{{- /* compute annotations */}}
{{- $annotations := include "application.version.annotations" . | mustFromJson }}

{{- /* compose into metadata */}}
{{- $metadata := (dict) }}
{{- $metadata := set $metadata "name" .VERSION_NAME }}
{{- $metadata := set $metadata "namespace" .VERSION_NAMESPACE }}
{{- $metadata := set $metadata "labels" $labels }}
{{- $metadata := set $metadata "annotations" $annotations }}

apiVersion: v1
kind: Service
{{- if .serviceSpecification }}
metadata:
{{- if .serviceSpecification.metatdata }}
  {{ toYaml (merge .serviceSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .serviceSpecification.metatdata */}}
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
{{- end }} {{- /* define "env.deployment-istio.version.service" */}}
