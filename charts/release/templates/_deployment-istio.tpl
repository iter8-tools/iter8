{{- define "env.deployment-istio" }}

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

{{- /* Service */}}
{{ include "env.deployment-istio.service" . }}
---

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.deployment-istio.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.deployment-istio.canary" . }}
{{- else if eq "mirror" .Values.application.strategy }}
{{ include "env.deployment-istio.mirror" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.deployment-istio" */}}