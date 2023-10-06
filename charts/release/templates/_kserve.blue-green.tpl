{{- define "env.kserve.blue-green" }}

{{- /* prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions" . | mustFromJson }}

{{- /* weight-config ConfigMaps */}}
{{- range $i, $v := $versions }}
{{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap */}}
{{ include "env.kserve.blue-green.routemap" . }}

{{- end }} {{- /* define "env.kserve.blue-green" */}}