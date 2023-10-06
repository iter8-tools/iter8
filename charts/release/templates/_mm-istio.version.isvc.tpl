{{- define "env.mm-istio.version.isvc" }}

{{- /* compute basic metadata */}}
{{- $metadata := include "application.version.metadata" . | mustFromJson }}
{{- /* add annotation serving.kserve.io/deploymentMode */}}
{{- $metadata := set $metadata "annotations" (merge $metadata.annotations (dict "serving.kserve.io/deploymentMode" "ModelMesh")) }}

{{- /* define InferenceServcie */}}
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
{{- if not .modelFormat }} {{- /* require .modelFormat */}}
{{- print "missing field: modelFormat required when inferenceServiceSpecification absent" | fail }}
{{- end }} {{- /* if not .modelFormat */}}
{{- if not .storageUri }} {{- /* require .storageUri */}}
{{- print "missing field: storageUri required when inferenceServiceSpecification absent" | fail }}
{{- end }} {{- /* if not .storageUri */}}
metadata:
  {{ toYaml $metadata | nindent 2 | trim }}
spec:
  predictor:
    model:
      modelFormat:
        name: {{ .modelFormat }}
      storageUri: {{ .storageUri }}
{{- end }} {{- /* if .inferenceServiceSpecification */}}

{{- end }} {{- /* define "env.mm-istio.version.isvc" */}}
