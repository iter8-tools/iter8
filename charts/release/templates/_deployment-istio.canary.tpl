{{- define "env.deployment-istio.canary" }}

{{- /* routemap */}}
{{ include "env.deployment-istio.canary.routemap" . }}

{{- end }} {{- /* define "env.deployment-istio.canary" */}}