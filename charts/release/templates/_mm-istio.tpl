{{- define "env.mm-istio" }}

{{- /* Prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions.kserve-mm" . | mustFromJson }}

{{- /* InferenceServices */}}
{{- range $i, $v := $versions }}
{{ include "env.mm-istio.version.isvc" $v }}
---
{{- end }} {{- /* range $i, $v := $versions */}}

{{- /* ServiceEntry */}}
{{ include "env.mm-istio.service" . }}
---

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.mm-istio.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.mm-istio.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.mm-istio.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.mm-istio.canary" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.mm-istio" */}}