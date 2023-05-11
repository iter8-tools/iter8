{{- define "iter8-traffic-template.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "iter8-traffic-template.labels" -}}
  labels:
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}

{{- define "resolve.modelVersions" }}
{{- /* produce a list of versions with all fields filled in with user specified values or defaults */}}
  {{- /* default values for fields depend on trafficStrategy */}}
  {{- $defaultNamespace := "modelmesh-serving" }}
  {{- $defaultWeight := ternary "100" "50" (eq .Values.trafficStrategy "mirror") }}
  {{- $defaultMatch := ternary (list (dict "headers" (dict "traffic" (dict "exact" "test")))) (dict) (eq .Values.trafficStrategy "canary") }}

  {{- $mV := list }}
  {{- if .Values.modelVersions }}
    {{- /* modelVersions were listed so just fill in any missing fields */}}
    {{- range $i, $ver := .Values.modelVersions }}
      {{- $v := merge $ver }}
      {{- $v = set $v "name" (default (printf "%s-%d" $.Values.modelName $i) $ver.name) }}
      {{- $v = set $v "namespace" (default $defaultNamespace $ver.namespace) }}
      {{- $v = set $v "weight" (default $defaultWeight $ver.weight | toString) }}
      {{- $v = set $v "match" (default $defaultMatch $ver.match) }}
      {{- $mV = append $mV $v }}
    {{- end }}
  {{- else }} {{- /* modelVersions NOT set, so use defaults for all fields  */}}
    {{- $mV = append $mV (dict "name" (printf "%s-0" .Values.modelName) "namespace" $defaultNamespace "weight" $defaultWeight "match" $defaultMatch ) }}
    {{- $mV = append $mV (dict "name" (printf "%s-1" .Values.modelName ) "namespace" $defaultNamespace "weight" $defaultWeight "match" $defaultMatch ) }}
  {{- end }}
  {{- mustToJson $mV }}
{{- end }}
