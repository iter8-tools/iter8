{{- define "release.labels" -}}
  labels:
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "application.name" -}}
{{- if (and .Values.application .Values.application.metadata .Values.application.metadata.name) -}}
{{ .Values.application.metadata.name -}}
{{- else -}}
{{ .Release.Name }}
{{- end -}}
{{- end -}} {{- /* define "application.name" */}}

{{- define "application.namespace" -}}
{{- if (and .Values.application .Values.application.metadata .Values.application.metadata.namespace) -}}
{{ .Values.application.metadata.namespace -}}
{{- else -}}
{{ .Release.Namespace }}
{{- end -}}
{{- end -}} {{- /* define "application.namespace" */}}

{{- define "application.version.labels" -}}
{{- $labels := (dict "iter8.tools/watch" "true") }}
{{- if (and .metadata .metadata.labels) }}
{{ $labels := merge $labels .metadata.labels }}
{{- end }}
{{- if (and .application.metadata .application.metadata.labels) }}
{{ $labels := merge $labels .application.metadata.labels }}
{{- end }}
{{- /* return as JSON */}}
{{- mustToJson $labels }}
{{- end -}} {{- /* define "application.version.labels" */}}

{{- define "application.version.annotations" -}}
{{- $annotations := (dict) }}
{{- if (and .metadata .metadata.annotations) }}
{{ $annotations := merge $annotations .metadata.annotations }}
{{- end }}
{{- if (and .application.metadata .application.metadata.annotations) }}
{{ $annotations := merge $annotations .application.metadata.annotations }}
{{- end }}
{{- /* return as JSON */}}
{{- mustToJson $annotations }}
{{- end -}} {{- /* define "application.version.annotations" */}}

{{- define "routemap.metadata" }}
metadata:
  name: {{ template "application.name" . }}-routemap
  namespace: {{ template "application.namespace" . }}
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: {{ .Values.iter8Version }}
{{- end }} {{- /* define "routemap.metadata" */}}

{{- define "normalize.versions" }}
  {{- /* Assumption: .Values.application is valid */}}
  {{- $metadata := dict }}
  {{- if .Values.application.metadata }}
  {{- $metadata := merge $metadata .Values.application.metadata }}
  {{- end }} {{- /* if .Values.application.metadata */}}

  {{- $APP_NAME := .Release.Name }}
  {{- if $metadata.name }}
  {{- $APP_NAME := $metadata.name }}
  {{- end }}
  {{- $APP_NAMESPACE := .Release.Namespace }}
  {{- if $metadata.namespace }}
  {{- $APP_NAMESPACE := $metadata.namespace }}
  {{- end }}

  {{- $defaultMatch := ternary (list (dict "headers" (dict "traffic" (dict "exact" "test")))) (dict) (eq .Values.application.strategy "canary") }}

  {{- $normalizedVersions := list }}
  {{- range $i, $v := .Values.application.versions -}}
    {{- $version := merge $v }}
    {{- $VERSION_NAME := (printf "%s-%d" $APP_NAME $i) }}
    {{- if (and $v.metadata $v.metadata.name) }}
    {{- $VERSION_NAME := $v.metadata.name }}
    {{- end }} {{- /* if (and $v.metadata $.vmetadata.name) */}}
    {{- $VERSION_NAMESPACE := $APP_NAMESPACE }}
    {{- if (and $v.metadata $v.metadata.namespace) }}
    {{- $VERSION_NAMESPACE := $v.metadata.namespace }}
    {{- end }} {{- /* if (and $v.metadata $.vmetadata.namespace) */}}
    {{- $version := merge $v }}
    {{- $version = set $version "VERSION_NAME" $VERSION_NAME -}}
    {{- $version = set $version "VERSION_NAMESPACE" $VERSION_NAMESPACE -}}

    {{- $application := (dict) }}
    {{- $application := set $application "APP_NAME" $APP_NAME }}
    {{- $application := set $application "APP_NAMESPACE" $APP_NAMESPACE }}
    {{- $application := set $application "metadata" $metadata }}
    {{- $version := set $version "application" $application}}

    {{- $version = set $version "weight" (default 50 $version.weight | toString) }}
    {{- $version = set $version "match" (default $defaultMatch $version.match) }}

    {{- $normalizedVersions = append $normalizedVersions $version }}
  {{- end }} {{- /* range $i, $v := .Values.application.versions */}}
  {{- mustToJson $normalizedVersions }}
{{- end }} {{- /* define "normalize.versions" */}}

{{- define "kserve.host" -}}
{{- $APP_NAMESPACE := .Release.Namespace -}}
{{- if (and .Values.application .Values.application.metadata) -}}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace -}}
{{- end -}}
{{- if eq "kserve-0.10" .Values.environment -}}
predictor-default.{{ $APP_NAMESPACE }}.svc.cluster.local
{{- else }} {{- /* kserve-0.11 or kserve */ -}}
predictor.{{ $APP_NAMESPACE }}.svc.cluster.local
{{- end }} {{- /* if eq ... .Values.environment */ -}}
{{- end }} {{- /* define "kserve.host" */ -}}

{{- define "mm.serviceHost" -}}
{{- $host := "modelmesh-serving.modelmesh-serving.svc.cluster.local" -}}
{{- if (and .Values.service .Values.service.host) -}}
{{- $host := .Values.service.host -}}
{{- end -}} {{- /* if (and .Values.service .Values.service.host) */ -}}
{{ $host }}
{{- end }} {{- /* define "mm.servingHost" */ -}}

{{- define "mm.servicePort" -}}
{{- $port := 8033 -}}
{{- if (and .Values.service .Values.service.port) -}}
{{- $port := .Values.service.port -}}
{{- end -}} {{- /* if (and .Values.service .Values.service.port) */ -}}
{{ $port }}
{{- end }} {{- /* define "mm.servingPort" */ -}}
