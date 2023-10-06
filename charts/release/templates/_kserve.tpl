{{- define "env.kserve" }}

{{- /* Prepare versions for simpler processing */}}
{{- $versions := include "normalize.versions" . | mustFromJson }}

{{- range $i, $v := .Values.application.versions }}
{{- /* InferenceService */}}
{{ include "env.kserve.version.isvc" . }}
---
{{- end }} {{- /* range $i, $v := .Values.application.versions */}}

{{- /* Service */}}
{{ include "env.kserve.service" . }}
---

{{- /* routemap (and other strategy specific objects) */}}
{{- if not .Values.application.strategy }}
{{ include "env.kserve.none" . }}
{{- else if eq "none" .Values.application.strategy }}
{{ include "env.kserve.none" . }}
{{- else if eq "blue-green" .Values.application.strategy }}
{{ include "env.kserve.blue-green" . }}
{{- else if eq "canary" .Values.application.strategy }}
{{ include "env.kserve.canary" . }}
{{- end }} {{- /* if eq ... .Values.application.strategy */}}

{{- end }} {{- /* define "env.kserve" */}}