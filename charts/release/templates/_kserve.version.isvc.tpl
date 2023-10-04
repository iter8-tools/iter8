{{- define "env.kserve.version.isvc" }}

{{- $labels := (merge (dict "iter8.tools/watch" "true" "app" .VERSION_NAME) .metadata.labels) }}
{{- $metadata := (dict "name" .VERSION_NAME "namespace" .VERSION_NAMESPACE "labels" $labels) }}

apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
{{- if .inferenceServiceSpecification }}
metadata:
{{- if .inferenceServiceSpecification.metatdata }}
  {{ toYaml (merge .inferenceServiceSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .inferenceServiceSpecification.metatdata */}}
spec:
  {{ toYaml .inferenceServiceSpecification.spec | nindent 2  | trim }}
{{- else }}
{{- if not .runtime }} {{- /* require .runtime */}}
{{- print "missing field: runtime required when inferenceServiceSpecification absent" | fail }}
{{- end }} {{- /* if not .runtime */}}
{{- if not .storageUri }} {{- /* require .storageUri */}}
{{- print "missing field: storageUri required when inferenceServiceSpecification absent" | fail }}
{{- end }} {{- /* if not .storageUri */}}
metadata:
  {{ toYaml $metadata | nindent 2 | trim }}
spec:
  predictor:
    minReplicas: 1
    model:
      modelFormat:
        name: {{ .modelFormat }}
      runtime: {{ .runtime }}
      storageUri: {{ .storageUri }}
{{- end }} {{- /* if .inferenceServiceSpecification */}}
{{- end }} {{- /* define "env.kserve.isvc" */}}
