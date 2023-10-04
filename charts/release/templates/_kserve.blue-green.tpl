{{- define "env.kserve.blue-green" }}
{{- $versions := include "normalize.versions" . | mustFromJson }}
{{- range $i, $v := $versions }}
{{ include "configmap.weight-config" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}
{{ include "env.kserve.blue-green.routemap" . }}
{{- end }} {{- /* define "env.kserve.blue-green" */}}