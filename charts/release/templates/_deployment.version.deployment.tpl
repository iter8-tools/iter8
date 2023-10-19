{{- define "env.deployment.version.deployment" }}

{{- /* compute basic metadata */}}
{{- $metadata := include "application.version.metadata" . | mustFromJson }}

apiVersion: apps/v1
kind: Deployment
{{- if .deploymentSpecification }}
metadata:
{{- if .deploymentSpecification.metadata }}
  {{ toYaml (merge .deploymentSpecification.metadata $metadata) | nindent 2 | trim }}
{{- else }}
  {{ toYaml $metadata | nindent 2 | trim }}
{{- end }} {{- /* if .deploymentSpecification.metadata */}}
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

{{- end }} {{- /* define "env.deployment.version.deployment" */}}
