{{- define "env.mm-istio.canary" }}

{{- /* routemap */}}
{{ include "env.mm-istio.canary.routemap" . }}

{{- end }} {{- /* define "env.mm-istio.canary" */}}