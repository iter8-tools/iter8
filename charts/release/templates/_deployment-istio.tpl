{{- define "env.deployment-istio" }}

{{- include "default.gateway" . }}
---

{{- $APP_NAME := .Release.Name }}
{{- $APP_NAMESPACE := .Release.Namespace }}
{{- if (and .Values.application .Values.application.metadata) }}
{{- $APP_NAME := .Values.application.metadata.name }}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
{{- end }}

{{- if not .Values.application }}
  {{- printf "No application versions specified" | fail }}
{{- end }} {{- /* if not .Values.application */}}

{{- range $i, $v := .Values.application.versions }}
{{- $VERSION_NAME := default (printf "%s-%d" $APP_NAME $i) $v.metadata.name }}
{{- $VERSION_NAMESPACE := default $APP_NAMESPACE $v.metadata.namespace }}
{{ $version := merge $v (dict "APP_NAME" $APP_NAME "VERSION_NAME" $VERSION_NAME "VERSION_NAMESPACE" $VERSION_NAMESPACE) }}

{{ include "env.deployment-istio.version.deployment" . }}
---
{{ include "env.deployment-istio.version.service" . }}
---
{{- end }} {{- /* range $i, $v := .Values.application.versions */}}

{{ include "env.deployment-istio.service" . }}
---

{{- if eq "none" .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.deployment-istio.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.deployment-istio.canary" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.deployment-istio" */}}