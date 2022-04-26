{{- define "task.readiness" -}}
{{- /* Apply only if user has specified resources to check for readiness */ -}}
{{- if .Values.ready }}
  {{- /* Review each entry in the .Values.ready map */ -}}
  {{- /* Create a task for each type of resource specified; only one resource of each type can be specified */ -}}
  {{- /* Note that "timeout" and "namespace" are special fields of the map; not resource types */ -}}
  {{- range $key, $value := .Values.ready }}
    {{- /* Filter out blacklisted fields "timeout" and "namespace" */ -}}
    {{- if not (has $key (list "namespace" "timeout")) }}
      {{- /* Before creating the task, verify that we have a definition for the readiness check */ -}}
      {{- /* Such definitions are in the map readinessDefintions */ -}}
      {{- if not (hasKey $.Values.iter8lib.readinessDefinitions $key) }}
        {{- fail (print "No readinessDefinition for resource type " $key) }}
      {{- end }}
      {{- /* Prepare arguments to task before creation */ -}}
      {{- $taskArgs := dict "name" $value }}
      {{- if $.Values.ready.namespace }}
        {{- $_ := set $taskArgs "namespace" $.Values.ready.namespace }}
      {{- end }}
      {{- if $.Values.ready.timeout }}
        {{- $_ := set $taskArgs "timeout" $.Values.ready.timeout }}
      {{- end }}
      {{- $_ := mustMerge $taskArgs (get $.Values.iter8lib.readinessDefinitions $key) }}
      {{- /* Finally, write the task */ -}}
# task: check if Kubernetes resource exists and if "ready"
- task: k8s-object-ready
  with:
{{ toYaml $taskArgs | indent 4 }}
    {{- end }}
  {{ end }}
{{- end }}
{{ end }}
