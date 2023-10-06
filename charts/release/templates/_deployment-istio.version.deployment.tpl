{{- define "env.deployment-istio.version.deployment" }}

{{- /* compute labels */}}
{{- $labels := include "application.version.labels" . | mustFromJson }}

{{- /* compute annotations */}}
{{- $annotations := include "application.version.annotations" . | mustFromJson }}

{{- /* compose into metadata */}}
{{- $metadata := (dict) }}
{{- $metadata := set $metadata "name" .VERSION_NAME }}
{{- $metadata := set $metadata "namespace" .VERSION_NAMESPACE }}
{{- $metadata := set $metadata "labels" $labels }}
{{- $metadata := set $metadata "annotations" $annotations }}

apiVersion: apps/v1
kind: Deployment
{{- if .deploymentSpecification }}
metadata:
{{- if .deploymentSpecification.metatdata }}
  {{ toYaml (merge .deploymentSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .deploymentSpecification.metatdata */}}
spec:
  {{ toYaml .deploymentSpecification.spec | nindent 2  | trim }}
{{- else }}
{{- if not .image }} {{- /* require .image */}}
{{- print "missing field: image required when deploymentSpecification absent" | fail }}
{{- end }} {{- /* if not .image */}}
{{- if not .port }} {{- /* require .port */}}
{{- print "missing field: port required when deploymentSpecification absent" | fail }}
{{- end }} {{- /* if not .port */}}
metadata:
  {{ toYaml $metadata | nindent 2 | trim }}
spec:
  selector:
    matchLabels:
      app: {{ .VERSION_NAME }}
  template:
    metadata:
      labels:
        app: {{ .VERSION_NAME }}
    spec:
      containers:
      - name: {{ .VERSION_NAME }}
        image: {{ .image }}
        ports:
        - containerPort: {{ .port }}
{{- end }} {{- /* if .deploymentSpecification */}}

{{- end }} {{- /* define "env.deployment-istio.version.deployment" */}}
