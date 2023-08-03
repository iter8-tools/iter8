{{- define "routemap-mirror" }}
{{- $versions := include "resolve.appVersions" . | mustFromJson }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.appName }}-routemap
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: {{ .Values.iter8Version }}
data:
  strSpec: |
    versions: 
{{- range $i, $v := $versions }}
    - weight: {{ $v.weight }}
      resources:
      {{- if gt $i 0 }}
      - gvrShort: cm
        name: {{ $v.name }}-weight-config
        namespace: {{ $v.namespace }}
      {{- end }}
      - gvrShort: isvc
        name: {{ $v.name }}
        namespace: {{ $v.namespace }}
{{- end }}
    routingTemplates:
      {{ .Values.strategy }}:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: {{ .Values.appName }}
          spec:
            gateways:
            - {{ .Values.externalGateway }}
            - mesh
            hosts:
            - {{ .Values.appName }}.{{ .Release.Namespace }}
            - {{ .Values.appName }}.{{ .Release.Namespace }}.svc
            - {{ .Values.appName }}.{{ .Release.Namespace }}.svc.cluster.local
            http:
            - route:
              - destination:
                  host: {{ .Values.modelmeshServingService }}.{{ .Release.Namespace }}.svc.cluster.local
                  port:
                    number: {{ $.Values.modelmeshServingPort }}
                headers:
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions 0).name }}"
                  response:
                    add:
                      mm-vmodel-id: "{{ (index $versions 0).name }}"
              {{ `{{- if gt (index .Weights ` }} 1 {{ `) 0 }}`}}
              mirror:
                host: {{ .Values.modelmeshServingService }}.{{ .Release.Namespace }}.svc.cluster.local
                port:
                  number: {{ $.Values.modelmeshServingPort }}
              mirrorPercentage:
                value: {{ `{{ index .Weights `}} 1 {{` }}`}}
              headers:
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions 1).name }}"
                  response:
                    add:
                      mm-vmodel-id: "{{ (index $versions 0).name }}"
              {{ `{{- end }}`}}
immutable: true
{{- end }}