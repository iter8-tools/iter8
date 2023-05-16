{{- define "routemap-canary" }}
{{- $versions := include "resolve.modelVersions" . | mustFromJson }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Values.modelName }}-routemap
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: {{ .Values.iter8Version }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: isvc
        name: {{ default (printf "%s-%d" $.Values.modelName $i) $v.name }}
        namespace: {{ default "modelmesh-serving" $v.namespace }}
    {{- end }}
    routingTemplates:
      {{ .Values.trafficStrategy }}:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: {{ .Values.modelName }}
          spec:
            gateways:
            - {{ .Values.externalGateway }}
            - mesh
            hosts:
            - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}
            - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}.svc
            - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}.svc.cluster.local
            http:
            {{- /* For candidate versions, ensure mm-model header is required in all matches */}}
            {{- range $i, $v := (rest $versions) }}
            {{- /* continue only if candidate is ready (weight > 0) */}}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
            - match:
              {{- /* A match may have several ORd clauses */}}
              {{- range $j, $m := $v.match }}
              {{- /* include any other header requirements */}}
              {{- if (hasKey $m "headers") }}
              - headers:
{{ toYaml (pick $m "headers").headers | indent 18 }}
                {{- end }}
                {{- /* include any other (non-header) requirements */}}
                {{- if gt (omit $m "headers" | keys | len) 0 }}
{{ toYaml (omit $m "headers") | indent 16 }}
                {{- end }}
              {{- end }}
              route:
              - destination:
                  host: {{ $.Values.modelmeshServingService }}.{{ $.Values.modelmeshServingNamespace }}.svc.cluster.local
                  port:
                    number: {{ $.Values.modelmeshServingPort }}
                headers:
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions (add1 $i)).name }}"
            {{ `{{- end }}`}}
            {{- end }}
            - route:
              - destination:
                  host: {{ $.Values.modelmeshServingService }}.{{ $.Values.modelmeshServingNamespace }}.svc.cluster.local
                  port:
                    number: {{ $.Values.modelmeshServingPort }}
                headers:
                  request:
                    set:
                      mm-vmodel-id: "{{ (index $versions 0).name }}"
immutable: true
{{- end }}
