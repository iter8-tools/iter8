{{- define "env.deployment" }}

{{- /* Prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}

{{- range $i, $v := $versions }}

{{- /* Deployment */}}
{{ include "env.deployment.version.deployment" $v }}
---
{{- /* Service */}}
{{ include "env.deployment.version.service" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else }}
{{- printf "unknown or invalid application strategy (%s) for environment (%s)" .Values.application.strategy .Values.environment | fail }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.deployment" */}}