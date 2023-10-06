{{- define "env.deployment-istio" }}

{{- include "default.gateway" . }}
---

{{- /* Prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions" . | mustFromJson }}

{{- range $i, $v := $versions }}
{{- /* Deployment */}}
{{ include "env.deployment-istio.version.deployment" . }}
---
{{- /* Service */}}
{{ include "env.deployment-istio.version.service" . }}
---
{{- end }} {{- /* range $i, $v := .Values.application.versions */}}

{{- /* Service */}}
{{ include "env.deployment-istio.service" . }}
---

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.mm-istio.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.deployment-istio.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.deployment-istio.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.deployment-istio.canary" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.deployment-istio" */}}