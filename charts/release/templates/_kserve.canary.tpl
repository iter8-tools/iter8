{{- define "env.kserve.canary" }}

{{- /* routemap */}}
{{ include "env.kserve.canary.routemap" . }}

{{- end }} {{- /* define "env.kserve.canary" */}}