{{- define "env.deployment-gtw.blue-green.routemap" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}
{{- $APP_PORT := pluck "port" (dict "port" 80) $.Values.application | first }}

apiVersion: v1
kind: ConfigMap
{{- template "routemap.metadata" . }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: svc
        name: {{ template "svc.name" $v }}
        namespace: {{ template "svc.namespace" $v }}
      - gvrShort: deploy
        name: {{ template "deploy.name" $v }}
        namespace: {{ template "deploy.namespace" $v }}
      - gvrShort: cm
        name: {{ $v.VERSION_NAME }}-weight-config
        namespace: {{ $v.VERSION_NAMESPACE }}
      weight: {{ $v.weight }}
    {{- end }} {{- /* range $i, $v := $versions */}}
    routingTemplates:
      {{ .Values.application.strategy }}:
        gvrShort: httproute
        template: |
          apiVersion: gateway.networking.k8s.io/v1beta1
          kind: HTTPRoute
          metadata:
            name: {{ $APP_NAME }}
            namespace: {{ $APP_NAMESPACE }}
          spec:
            hostnames:
            - {{ $APP_NAME }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
            parentRefs:
            - group: ""
              kind: Service
              name: {{ $APP_NAME }}
              port: {{ $APP_PORT }}
            {{- if .Values.gateway }}
            - name: {{ .Values.gateway }}
            {{- end }}
            rules:
            - backendRefs:
              {{- range $i, $v := $versions }}
              - group: ""
                kind: Service
                name: {{ template "svc.name" $v }}
                port: {{ $v.port }}
                {{- if gt (len $versions) 1 }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights ` }}{{ print $i }}{{ ` }}` }}
                {{ `{{- end }}` }}
                {{- end  }}
                filters:
                - type: ResponseHeaderModifier
                  responseHeaderModifier:
                    add:
                    - name: app-version
                      value: {{ template "svc.name" $v }}
              {{- end }} {{- /* range $i, $v := $versions */}}
{{- end }} {{- /* define "env.deployment-gtw.blue-green.routemap" */}}
