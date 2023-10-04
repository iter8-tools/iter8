{{- define "release.labels" -}}
  labels:
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "normalize.versions" }}
  {{- $APP_NAME := .Release.Name }}
  {{- $APP_NAMESPACE := .Release.Namespace }}
  {{- if (and .Values.application .Values.application.metadata) }}
  {{- $APP_NAME := .Values.application.metadata.name }}
  {{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
  {{- end }}
  {{- $normalizedVersions := list }}
  {{- range $i, $v := .Values.application.versions -}}
    {{- $version := merge $v }}
    {{- $version = set $version "APP_NAME" $APP_NAME -}}
    {{- $version = set $version "APP_NAMESPACE" $APP_NAMESPACE -}}
    {{- $version = set $version "VERSION_NAME" (default (printf "%s-%d" $APP_NAME $i) $version.metadata.name) -}}
    {{- $version = set $version "VERSION_NAMESPACE" (default $APP_NAMESPACE $version.metadata.namespace) -}}
    {{- $version = set $version "weight" (default 50 $version.weight | toString) }}
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
