{{- define "routemap-bluegreen-kserve" }}
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
      {{ .Values.trafficStrategy }}:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: {{ .Values.modelName }}
          spec:
            gateways:
            - knative-serving/knative-ingress-gateway
            - knative-serving/knative-local-gateway
            - mesh
            hosts:
            - {{ .Values.modelName }}.{{ .Release.Namespace }}
            - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc
            - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc.cluster.local
            http:
            - name: {{ (index $versions 0).name }}
              match:
              - headers:
                  branch:
                    exact: {{ (index $versions 0).name }}
              rewrite:
                uri: /v2/models/{{ (index $versions 0).name }}/infer
              route:
              # primary model
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers: 
                  request:
                    set:
                      Host: {{ (index $versions 0).name }}-predictor-default.{{ .Release.Namespace }}.svc.cluster.local
                    remove: 
                    - branch
                  response:
                    add:
                      mm-vmodel-id: {{ (index $versions 0).name }}
            # other models
            {{- range $i, $v := (rest $versions) }}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
            - name: {{ (index $versions (add1 $i)).name }}
              match:
              - headers:
                  branch:
                    exact: {{ (index $versions (add1 $i)).name }}
              rewrite:
                uri: /v2/models/{{ (index $versions (add1 $i)).name }}/infer
              route:
              # primary model
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers: 
                  request:
                    set:
                      Host: {{ (index $versions (add1 $i)).name }}-predictor-default.{{ $.Release.Namespace }}.svc.cluster.local
                    remove: 
                    - branch
                  response:
                    add:
                      mm-vmodel-id: {{ (index $versions (add1 $i)).name }}
            {{ `{{- end }}`}}     
            {{- end }}
            - name: split
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}`}}
                headers:
                  request:
                    set:
                      branch: {{ (index $versions 0).name }}
                      host: {{ .Values.modelName }}.{{ .Release.Namespace }}
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                weight: {{ `{{ index .Weights `}}{{ print (add1 $i) }}{{` }}`}}
                headers:
                  request:
                    set:
                      branch: {{ (index $versions (add1 $i)).name }}
                      host: {{ $.Values.modelName }}.{{ $.Release.Namespace }}
              {{ `{{- end }}`}}
              {{- end }}
immutable: true
{{- end }}
