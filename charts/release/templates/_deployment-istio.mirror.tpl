{{- define "env.deployment-istio.mirror" }}

{{- /* prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}

{{- /* weight-config ConfigMaps except for primary */}}
{{- range $i, $v := (rest $versions) }}
    {{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* routemap */}}
{{ include "env.deployment-istio.mirror.routemap" . }}

{{- end }} {{- /* define "env.deployment-istio.mirror" */}}