{{- define "env.deployment-istio.mirror.routemap" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- $versions := include "normalize.versions.deployment" . | mustFromJson }}

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
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: {{ $APP_NAME }}
            namespace: {{ $APP_NAMESPACE }}
          spec:
            gateways:
            {{- if .Values.gateway }}
            - {{ .Values.gateway }}
            {{- end }}
            - mesh
            hosts:
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
            http:
            - name: {{ $APP_NAME }}
              route:
              # primary version
              {{- $v := (index $versions 0) }}
              - destination:
                  host: {{ template "svc.name" $v }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                  port:
                    number: {{ $v.port }}
                {{- if gt (len $versions) 1 }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}` }}
                {{- end  }}
                headers: 
                  response:
                    add:
                      app-version: {{ template "svc.name" $v }}
              # other versions
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}` }}
              mirror:
                host: {{ template "svc.name" $v }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                port:
                  number: {{ $v.port }}
              mirrorPercentage:
                value: {{ `{{ index .Weights ` }}{{ print (add1 $i) }}{{ ` }}` }}
              {{ `{{- end }}` }}
              {{- end }}
{{- end }} {{- /* define "env.deployment-istio.mirror.routemap" */}}
