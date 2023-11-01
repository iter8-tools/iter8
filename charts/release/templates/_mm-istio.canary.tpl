{{- define "env.mm-istio.canary" }}

{{- /* ServiceEntry */}}
{{ include "env.mm-istio.service" . }}
---

{{- /* routemap */}}
{{ include "env.mm-istio.canary.routemap" . }}

{{- end }} {{- /* define "env.mm-istio.canary" */}}