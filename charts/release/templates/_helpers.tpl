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
{{ $labels = merge $labels .metadata.labels }}
{{- end }}
{{- if (and .application.metadata .application.metadata.labels) }}
{{ $labels = merge $labels .application.metadata.labels }}
{{- end }}
{{- /* return as JSON */}}
{{- mustToJson $labels }}
{{- end -}} {{- /* define "application.version.labels" */}}

{{- define "application.version.annotations" -}}
{{- $annotations := (dict) }}
{{- if (and .metadata .metadata.annotations) }}
{{ $annotations = merge $annotations .metadata.annotations }}
{{- end }}
{{- if (and .application.metadata .application.metadata.annotations) }}
{{ $annotations = merge $annotations .application.metadata.annotations }}
{{- end }}
{{- /* return as JSON */}}
{{- mustToJson $annotations }}
{{- end -}} {{- /* define "application.version.annotations" */}}

{{- define "application.version.metadata" -}}
{{- $labels := (include "application.version.labels" . | mustFromJson) }}
{{- $annotations := (include "application.version.annotations" . | mustFromJson) }}
{{- /* compose into metadata */}}
{{- $metadata := (dict) }}
{{- $metadata := set $metadata "name" .VERSION_NAME }}
{{- $metadata := set $metadata "namespace" .VERSION_NAMESPACE }}
{{- $metadata := set $metadata "labels" $labels }}
{{- $metadata := set $metadata "annotations" $annotations }}
{{- /* return as JSON */}}
{{- mustToJson $metadata }}
{{- end -}} {{- /* define "application.version.metadata" */}}

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
  {{- $metadata = merge $metadata .Values.application.metadata }}
  {{- end }} {{- /* if .Values.application.metadata */}}

  {{- $APP_NAME := (include "application.name" .) }}
  {{- $APP_NAMESPACE := (include "application.namespace" .) }}

  {{- /* set default match for canary use cases: "traffic: test" */}}
  {{- /* this is arbitrary but enables trying it out quickly " */}}
  {{- $defaultMatch := ternary (list (dict "headers" (dict "traffic" (dict "exact" "test")))) (dict) (eq .Values.application.strategy "canary") }}

  {{- $normalizedVersions := list }}
  {{- range $i, $v := .Values.application.versions -}}
    {{- $version := merge $v }}
    {{- $VERSION_NAME := (printf "%s-%d" $APP_NAME $i) }}
    {{- if (and $v.metadata $v.metadata.name) }}
    {{- $VERSION_NAME = $v.metadata.name }}
    {{- end }} {{- /* if (and $v.metadata $v.metadata.name) */}}
    {{- $VERSION_NAMESPACE := $APP_NAMESPACE }}
    {{- if (and $v.metadata $v.metadata.namespace) }}
    {{- $VERSION_NAMESPACE = $v.metadata.namespace }}
    {{- end }} {{- /* if (and $v.metadata $v.metadata.namespace) */}}
    {{- $version := merge $v }}
    {{- $version = set $version "VERSION_NAME" $VERSION_NAME -}}
    {{- $version = set $version "VERSION_NAMESPACE" $VERSION_NAMESPACE -}}

    {{- $application := (dict) }}
    {{- $application := set $application "APP_NAME" $APP_NAME }}
    {{- $application := set $application "APP_NAMESPACE" $APP_NAMESPACE }}
    {{- $application := set $application "metadata" $metadata }}
    {{- $version := set $version "application" $application }}

    {{- $version = set $version "weight" (default 50 $version.weight | toString) }}
    {{- $version = set $version "match" (default $defaultMatch $version.match) }}

    {{- $normalizedVersions = append $normalizedVersions $version }}
  {{- end }} {{- /* range $i, $v := .Values.application.versions */}}
  {{- mustToJson $normalizedVersions }}
{{- end }} {{- /* define "normalize.versions" */}}

{{- /* deployment specific wrapping for normalize.versions */}}
{{- define "normalize.versions.deployment" }}
{{- $versions := include "normalize.versions" . | mustFromJson }}
  {{- $normalizedVersions := list }}
  {{- range $i, $v := $versions -}}
    {{- $version := merge $v }}

    {{- $version = set $version "port" (pluck "port" (dict "port" 80) $.Values.application $v | last) }}
    {{- if (and $v.serviceSpecification $v.serviceSpecification.ports) }}
      {{- $version = set $version "port" (pluck "port" $v (index $v.serviceSpecification.ports 0) | last) }}
    {{- end }}

    {{- $normalizedVersions = append $normalizedVersions $version }}
  {{- end }} {{- /* range $i, $v := $versions */}}
  {{- mustToJson $normalizedVersions }}
{{- end }} {{- /* define "normalize.versions.kserve" */}}

{{- /* kserve specific wrapping for normalize.versions */}}
{{- define "normalize.versions.kserve" }}
{{- $versions := include "normalize.versions" . | mustFromJson }}
  {{- $normalizedVersions := list }}
  {{- range $i, $v := $versions -}}
    {{- $version := merge $v }}

    {{- $version = set $version "modelFormat" (pluck "modelFormat" $.Values.application $v | last) }}
    {{- $version = set $version "runtime" (pluck "runtime" $.Values.application $v | last) }}
    {{- $protocolVersion := (pluck "protocolVersion" $.Values.application $v | last) }}
    {{- if $protocolVersion }}
      {{- $version = set $version "protocolVersion" $protocolVersion }}
    {{- end }} {{- /* if $protocolVersion */}}
    {{- $ports := (pluck "ports" $.Values.application $v | last) }}
    {{- if $ports }}
      {{- $version = set $version "ports" $ports }}
    {{- end }} {{- /* if $ports */}}

    {{- $normalizedVersions = append $normalizedVersions $version }}
  {{- end }} {{- /* range $i, $v := $versions */}}
  {{- mustToJson $normalizedVersions }}
{{- end }} {{- /* define "normalize.versions.kserve" */}}

{{- /* kserve-modelmesh specific wrapping for normalize.versions */}}
{{- define "normalize.versions.kserve-mm" }}
{{- $versions := include "normalize.versions" . | mustFromJson }}
  {{- $normalizedVersions := list }}
  {{- range $i, $v := $versions -}}
    {{- $version := merge $v }}

    {{- $version = set $version "modelFormat" (pluck "modelFormat" $.Values.application $v | last) }}

    {{- $normalizedVersions = append $normalizedVersions $version }}
  {{- end }} {{- /* range $i, $v := $versions */}}
  {{- mustToJson $normalizedVersions }}
{{- end }} {{- /* define "normalize.versions.kserve-mm" */}}

{{- /* Identify the name of a Deployment object */ -}}
{{- define "deploy.name" -}}
{{- if (and .deploymentSpecification .deploymentSpecification.metadata .deploymentSpecification.metadata.name) -}}
{{ .deploymentSpecification.metadata.name }}
{{- else -}}
{{ .VERSION_NAME }}
{{- end -}}
{{- end }} {{- /* define "deploy.name" */ -}}

{{- /* Identify the namespace of a Deployment object */ -}}
{{- define "deploy.namespace" -}}
{{- if (and .deploymentSpecification .deploymentSpecification.metadata .deploymentSpecification.metadata.namespace) -}}
{{ .deploymentSpecification.metadata.namespace }}
{{- else -}}
{{ .VERSION_NAMESPACE }}
{{- end -}}
{{- end }} {{- /* define "deploy.namespace" */ -}}

{{- /* Identify the name of a Service object */ -}}
{{- define "svc.name" -}}
{{- if (and .serviceSpecification .serviceSpecification.metadata .serviceSpecification.metadata.name) -}}
{{ .serviceSpecification.metadata.name }}
{{- else -}}
{{ .VERSION_NAME }}
{{- end -}}
{{- end }} {{- /* define "svc.name" */ -}}

{{- /* Identify the namespace of a Service object */ -}}
{{- define "svc.namespace" -}}
{{- if (and .serviceSpecification .serviceSpecification.metadata .serviceSpecification.metadata.namespace) -}}
{{ .serviceSpecification.metadata.namespace }}
{{- else -}}
{{ .VERSION_NAMESPACE }}
{{- end -}}
{{- end }} {{- /* define "svc.namespace" */ -}}

{{- /* Identify the name of an InferenceService object */ -}}
{{- define "isvc.name" -}}
{{- if (and .inferenceServiceSpecification .inferenceServiceSpecification.metadata .inferenceServiceSpecification.metadata.name) -}}
{{ .inferenceServiceSpecification.metadata.name }}
{{- else -}}
{{ .VERSION_NAME }}
{{- end -}}
{{- end }} {{- /* define "isvc.name" */ -}}

{{- /* Identify the namespace of an InferenceService object */ -}}
{{- define "isvc.namespace" -}}
{{- if (and .inferenceServiceSpecification .inferenceServiceSpecification.metadata .inferenceServiceSpecification.metadata.namespace) -}}
{{ .inferenceServiceSpecification.metadata.namespace }}
{{- else -}}
{{ .VERSION_NAMESPACE }}
{{- end -}}
{{- end }} {{- /* define "isvc.namespace" */ -}}

{{- define "kserve.host" -}}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- if eq "kserve-0.10" .Values.environment -}}
predictor-default.{{ $APP_NAMESPACE }}.svc.cluster.local
{{- else }} {{- /* kserve-0.11 or kserve */ -}}
predictor.{{ $APP_NAMESPACE }}.svc.cluster.local
{{- end }} {{- /* if eq ... .Values.environment */ -}}
{{- end }} {{- /* define "kserve.host" */ -}}

{{- define "mm.serviceHost" -}}
{{- $host := "modelmesh-serving.modelmesh-serving.svc.cluster.local" -}}
{{- if (and .Values.service .Values.service.host) -}}
{{- $host = .Values.service.host -}}
{{- end -}} {{- /* if (and .Values.service .Values.service.host) */ -}}
{{ $host }}
{{- end }} {{- /* define "mm.servingHost" */ -}}

{{- define "mm.servicePort" -}}
{{- $port := 8033 -}}
{{- if (and .Values.service .Values.service.port) -}}
{{- $port = .Values.service.port -}}
{{- end -}} {{- /* if (and .Values.service .Values.service.port) */ -}}
{{ $port }}
{{- end }} {{- /* define "mm.servingPort" */ -}}

