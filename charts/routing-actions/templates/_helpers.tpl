{{- define "iter8-traffic-template.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "iter8-traffic-template.labels" -}}
  labels:
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "resolve.appVersions" }}
{{- /* produce a list of versions with all fields filled in with user specified values or defaults */}}
  {{- /* default values for fields depend on strategy */}}
  {{- $defaultNamespace := .Release.Namespace }}
  {{- $defaultWeight := ternary "100" "50" (eq .Values.strategy "mirror") }}
  {{- $defaultMatch := ternary (list (dict "headers" (dict "traffic" (dict "exact" "test")))) (dict) (eq .Values.strategy "canary") }}

  {{- $mV := list }}
  {{- if .Values.appVersions }}
    {{- /* appVersions were listed so just fill in any missing fields */}}
    {{- range $i, $ver := .Values.appVersions }}
      {{- $v := merge $ver }}
      {{- $v = set $v "name" (default (printf "%s-%d" $.Values.appName $i) $ver.name) }}
      {{- $v = set $v "namespace" (default $defaultNamespace $ver.namespace) }}
      {{- $v = set $v "weight" (default $defaultWeight $ver.weight | toString) }}
      {{- $v = set $v "match" (default $defaultMatch $ver.match) }}
      {{- $mV = append $mV $v }}
    {{- end }}
  {{- else }} {{- /* appVersions NOT set, so use defaults for all fields  */}}
    {{- $mV = append $mV (dict "name" (printf "%s-0" .Values.appName) "namespace" $defaultNamespace "weight" $defaultWeight "match" $defaultMatch ) }}
    {{- $mV = append $mV (dict "name" (printf "%s-1" .Values.appName ) "namespace" $defaultNamespace "weight" $defaultWeight "match" $defaultMatch ) }}
  {{- end }}
  {{- mustToJson $mV }}
{{- end }}

{{- define "kserve.host" }}
{{- if eq "kserve-0.10" .Values.appType -}}
predictor-default.{{ .Release.Namespace }}.svc.cluster.local
{{- else }} {{- /* kserve-0.11 or kserve */ -}}
predictor.{{ .Release.Namespace }}.svc.cluster.local
{{- end }}
{{- end }} {{- /* define "kserve.host" */ -}}
