{{- define "env.kserve.version.isvc" }}

{{- /* compute basic metadata */}}
{{- $metadata := include "application.version.metadata" . | mustFromJson }}
{{- /* add annotation serving.kserve.io/deploymentMode */}}

{{- /* define InferenceServcie */}}
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
{{- if .inferenceServiceSpecification }}
metadata:
{{- if .inferenceServiceSpecification.metadata }}
  {{ toYaml (merge .inferenceServiceSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .inferenceServiceSpecification.metadata */}}
spec:
  {{ toYaml .inferenceServiceSpecification.spec | nindent 2  | trim }}
{{- else }}
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
      {{- if .protocolVersion }}
      protocolVersion: {{ .protocolVersion }}
      {{- end  }} {{- /* if .protocolVersion */}}
      {{- if .ports }}
      ports: 
      {{ toYaml .ports | nindent 6 | trim }}
      {{- end  }} {{- /* if .ports */}}
{{- end }} {{- /* if .inferenceServiceSpecification */}}

{{- end }} {{- /* define "env.kserve.version.isvc" */}}
