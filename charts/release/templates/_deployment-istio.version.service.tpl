{{- define "env.deployment-istio.version.service" }}

{{- $labels := (merge (dict "iter8.tools/watch" "true" "app" .VERSION_NAME) .metadata.labels) }}
{{- $metadata := (dict "name" .VERSION_NAME "namespace" .VERSION_NAMESPACE "labels" $labels) }}

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
