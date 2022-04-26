{{- define "k.manifest" -}}
{{ include "k.spec.secret" . }}
---
{{ include "k.spec.role" . }}
---
{{ include "k.spec.rolebinding" . }}
---
{{ include "k.result.secret" . }}
---
{{ include "k.result.role" . }}
---
{{ include "k.result.rolebinding" . }}
{{- if not .Values.iter8lib.disable.job }}
---
{{ include "k.job" . }}
{{- end }}
{{- end }}
