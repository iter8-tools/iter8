{{- define "routemap-bluegreen" }}
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
      - gvrShort: cm
        name: {{ $v.name }}-weight-config
        namespace: {{ $v.namespace }}
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
              # primary model
              - destination:
                  host: {{ .Values.modelmeshServingService }}.{{ .Release.Namespace }}.svc.cluster.local
                  port:
                    number: {{ $.Values.modelmeshServingPort }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}`}}
                headers: 
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions 0).name }}" 
                  response:
                    add:
                      app-version: "{{ (index $versions 0).name }}"
              # other models
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
              - destination:
                  host: {{ $.Values.modelmeshServingService }}.{{ $.Release.Namespace }}.svc.cluster.local
                  port:
                    number: {{ $.Values.modelmeshServingPort }}
                weight: {{ `{{ index .Weights `}}{{ print (add1 $i) }}{{` }}`}}
                headers:
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions (add1 $i)).name }}"
                  response:
                    add:
                      app-version: "{{ (index $versions (add1 $i)).name }}"
              {{ `{{- end }}`}}     
              {{- end }}
immutable: true
{{- end }}
