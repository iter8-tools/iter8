{{- define "env.deployment-gtw.blue-green" }}

{{- /* prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}

{{- /* weight-config ConfigMaps */}}
{{- range $i, $v := $versions }}
    {{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap */}}
{{ include "env.deployment-gtw.blue-green.routemap" . }}

{{- end }} {{- /* define "env.deployment-gtw.blue-green" */}}