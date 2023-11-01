{{- define "env.deployment-gtw" }}

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
{{ include "env.deployment-gtw.service" . }}
---

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.deployment-gtw.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.deployment-gtw.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.deployment-gtw.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.deployment-gtw.canary" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.deployment-gtw" */}}