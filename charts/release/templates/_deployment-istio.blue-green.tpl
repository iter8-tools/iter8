{{- define "env.deployment-istio.blue-green" }}
{{- $versions := include "normalize.versions" . | mustFromJson }}
{{- range $i, $v := $versions }}
{{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}
{{ include "env.deployment-istio.blue-green.routemap" . }}
{{- end }} {{- /* define "env.deployment-istio.blue-green" */}}