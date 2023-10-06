{{- define "env.mm-istio.blue-green" }}

{{- /* prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions" . | mustFromJson }}

{{- /* weight-config ConfigMaps */}}
{{- range $i, $v := $versions }}
{{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap */}}
{{ include "env.mm-istio.blue-green.routemap" . }}

{{- end }} {{- /* define "env.mm-istio.blue-green" */}}