{{- define "task.ready" }}
{{- if .Values.ready }}
{{- $typesToCheck := omit .Values.ready "timeout" "namespace" }}
{{- range $type, $name := $typesToCheck }}
{{- $definition := get $.Values.resourceTypes $type }}
{{- if not $definition }}
{{- cat "no type definition for: " $type | fail }}
{{- else }}
# task: test for existence and readiness of a resource
- task: ready
  with:
    name: {{ $name | quote }}
    group: {{ get $definition "Group" | quote }}
    version: {{ get $definition "Version" }}
    resource: {{ get $definition "Resource" }}
    {{- if (hasKey $definition "conditions") }}
    conditions:
{{ toYaml (get $definition "conditions") | indent 4 }}
    {{- end }} {{- /* if (hasKey $definition "conditions") */}}
{{- $namespace := coalesce $.Values.ready.namespace $.Release.Namespace }}    
{{- if $namespace }}
    namespace: {{ $namespace }}
{{- end }} {{- /* if $namespace */}}
{{- if $.Values.ready.timeout }}
    timeout: {{ $.Values.ready.timeout }}
{{- end }} {{- /* if $.Values.ready.timeout */}}
{{- end }} {{- /* if not $definition */}}
{{- end }} {{- /* range $type, $name */}}
{{- end }} {{- /* {{- if .Values.ready */}}
{{- end }} {{- /* define "task.ready" */}}