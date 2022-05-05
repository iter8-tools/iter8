{{- define "k.manifest-istio" -}}
{{ include "k.job" . }}
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
