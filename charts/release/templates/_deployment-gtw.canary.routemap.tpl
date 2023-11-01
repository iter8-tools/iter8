{{- define "env.deployment-gtw.canary.routemap" }}

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
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
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
            # non-primary versions
            {{- range $i, $v := (rest $versions) }}
            - matches:
{{- toYaml $v.matches | nindent 14 }}
              backendRefs:
              - group: ""
                kind: Service
                name: {{ template "svc.name" $v }}
                port: {{ $v.port }}
                filters:
                - type: ResponseHeaderModifier
                  responseHeaderModifier:
                    add:
                    - name: app-version
                      value: {{ template "svc.name" $v }}
              {{- end }} {{- /* range $i, $v := (rest $versions) */}}
            # primary version (default)
            {{- $v := (index $versions 0) }}
            - backendRefs:
              - group: ""
                kind: Service
                name: {{ template "svc.name" $v }}
                port: {{ $v.port }}
                filters:
                - type: ResponseHeaderModifier
                  responseHeaderModifier:
                    add:
                    - name: app-version
                      value: {{ template "svc.name" $v }}
{{- end }} {{- /* define "env.deployment-gtw.canary.routemap" */}}
