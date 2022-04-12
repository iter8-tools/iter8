{{- define "k.manifest-cronjob" -}}
{{ include "k.cronjob" . }}
---
{{ include "k.spec.secret-istio" . }}
---
{{ include "k.spec.role" . }}
---
{{ include "k.spec.rolebinding" . }}
---
{{ include "k.result.role" . }}
---
{{ include "k.result.rolebinding" . }}
{{- end }}
