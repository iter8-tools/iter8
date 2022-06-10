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
---
{{- if eq "job" .Values.iter8lib.runner }}
{{ include "k.job" . }}
---
{{- end }}
{{- if eq "cronjob" .Values.iter8lib.runner }}
{{ include "k.cronjob" . }}
---
{{- end }}
{{ include "k.task.ready.rbac" . }}
{{- end }}