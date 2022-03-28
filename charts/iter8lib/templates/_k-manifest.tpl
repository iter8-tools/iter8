{{- define "k.manifest" -}}
{{- include "k.job" . }}
---
{{- include "k.spec.secret" . }}
---
{{- include "k.spec.role" . }}
---
{{- include "k.spec.rolebinding" . }}
---
{{- include "k.result.role" . }}
---
{{- include "k.result.rolebinding" . }}
{{- end -}}
