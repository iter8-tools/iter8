{{- define "env.mm-istio.blue-green" }}

{{- /* ServiceEntry */}}
{{ include "env.mm-istio.service" . }}
---

{{- /* prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions.kserve-mm" . | mustFromJson }}

{{- /* weight-config ConfigMaps */}}
{{- range $i, $v := $versions }}
{{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap */}}
{{ include "env.mm-istio.blue-green.routemap" . }}

{{- end }} {{- /* define "env.mm-istio.blue-green" */}}