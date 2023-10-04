{{- define "env.deployment-istio.blue-green.routemap" }}
{{- $APP_NAME := .Release.Name }}
{{- $APP_NAMESPACE := .Release.Namespace }}
{{- if (and .Values.application .Values.application.metadata) }}
{{- $APP_NAME := .Values.application.metadata.name }}
{{- $APP_NAMESPACE := .Values.application.metadata.namespace }}
{{- end }}
{{- $versions := include "normalize.versions" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $APP_NAME }}-routemap
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: {{ .Values.iter8Version }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: cm
        name: {{ $v.VERSION_NAME }}-weight-config
        namespace: {{ $v.VERSION_NAMESPACE }}
      - gvrShort: svc
        name: {{ $v.VERSION_NAME }}
        namespace: {{ $v.VERSION_NAMESPACE }}
      - gvrShort: deploy
        name: {{ $v.VERSION_NAME }}
        namespace: {{ $v.VERSION_NAMESPACE }}
      weight: {{ $v.weight }}
    {{- end }} {{- /* range $i, $v := .Values.application.versions */}}
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
            - {{ default "iter8-gateway" .Values.gateway }}
            - mesh
            hosts:
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
            http:
            - name: {{ (index $versions 0).VERSION_NAME }}
              route:
              # primary version
              - destination:
                  host: {{ (index $versions 0).VERSION_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                {{- if gt (len $versions) 1 }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}`}}
                {{- end  }}
                headers: 
                  response:
                    add:
                      app-version: {{ (index $versions 0).VERSION_NAME }}
              # other versions
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
              - destination:
                  host: {{ (index $versions (add1 $i)).VERSION_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                weight: {{ `{{ index .Weights `}}{{ print (add1 $i) }}{{` }}`}}
                headers:
                  response:
                    add:
                      app-version: {{ (index $versions (add1 $i)).VERSION_NAME }}
              {{ `{{- end }}`}}     
              {{- end }}
{{- end }} {{- /* define "env.deployment-istio.blue-green.routemap" */}}
