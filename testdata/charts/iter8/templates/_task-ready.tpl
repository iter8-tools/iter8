{{- define "task.ready.tn" }}
{{- if .Values.ready }}
{{- $namespace := coalesce .Values.ready.namespace .Release.Namespace }}    
{{- if $namespace }}
    namespace: {{ $namespace }}
{{- end }}
{{- if .Values.ready.timeout }}
    timeout: {{ .Values.ready.timeout }}
{{- end }}
{{- end }}
{{- end }}
{{- define "task.ready" }}
{{- if .Values.ready }}
{{- $namespace := coalesce .Values.ready.namespace .Release.Namespace }}
{{- if .Values.ready.service }}
# task: determine if Kubernetes Service exists
- task: ready
  with:
    name: {{ .Values.ready.service | quote }}
    version: v1
    resource: services
{{- include "task.ready.tn" . }}
{{- end }}
{{- if .Values.ready.deploy }}
# task: determine if Kubernetes Deployment exists and is Available
- task: ready
  with:
    name: {{ .Values.ready.deploy | quote }}
    group: apps
    version: v1
    resource: deployments
    condition: Available
{{- include "task.ready.tn" . }}
{{- end }}
{{- if .Values.ready.ksvc }}
# task: determine if Knative Service exists and is ready
- task: ready
  with:
    name: {{ .Values.ready.ksvc | quote }}
    group: serving.knative.dev
    version: v1
    resource: services
    condition: Ready
{{- include "task.ready.tn" . }}
{{- end }}
{{- if .Values.ready.isvc }}
# task: determine if KServe InferenceService exists and is ready
- task: ready
  with:
    name: {{ .Values.ready.isvc | quote }}
    group: serving.kserve.io
    version: v1beta1
    resource: inferenceservices
    condition: Ready
{{- include "task.ready.tn" . }}
{{- end }}
{{- if .Values.ready.chaosengine }}
# task: determine if chaos engine resource exists
- task: ready
  with:
    name: {{ .Values.ready.chaosengine | quote }}
    group: litmuschaos.io
    version: v1alpha1
    resource: chaosengines
{{- include "task.ready.tn" . }}
{{- end }}
{{- end }}
{{- end }}