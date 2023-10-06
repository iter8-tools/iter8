{{- define "env.mm-istio.version.isvc" }}

{{- /* compute labels */}}
{{- $labels := include "application.version.labels" . | mustFromJson }}

{{- /* compute annotations */}}
{{- $annotations := include "application.version.annotations" . | mustFromJson }}
{{- $annotations := merge (dict "serving.kserve.io/deploymentMode" "ModelMesh") $annotations }}

{{- /* compose into metadata */}}
{{- $metadata := (dict) }}
{{- $metadata := set $metadata "name" .VERSION_NAME }}
{{- $metadata := set $metadata "namespace" .VERSION_NAMESPACE }}
{{- $metadata := set $metadata "labels" $labels }}
{{- $metadata := set $metadata "annotations" $annotations }}

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
