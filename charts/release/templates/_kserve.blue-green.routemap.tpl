{{- define "env.kserve.blue-green.routemap" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- $versions := include "normalize.versions" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
{{ template "routemap.metadata" . }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: cm
        name: {{ $v.VERSION_NAME }}-weight-config
        namespace: {{ $v.VERSION_NAMESPACE }}
      - gvrShort: isvc
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
            - knative-serving/knative-ingress-gateway
            - knative-serving/knative-local-gateway
            - mesh
            hosts:
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
            http:
            # primary model version
            - name: {{ (index $versions 0).VERSION_NAME }}
              match:
              - headers:
                  branch:
                    exact: {{ (index $versions 0).VERSION_NAME }}
              rewrite:
                uri: /v2/models/{{ (index $versions 0).VERSION_NAME }}/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: {{ (index $versions 0).VERSION_NAME }}-{{ template "kserve.host" $ }}
                    remove:
                    - branch
                  response:
                    add:
                      app-version: {{ (index $versions 0).VERSION_NAME }}
            # other model versions
            {{- range $i, $v := (rest $versions) }}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
            - name: {{ (index $versions (add1 $i)).VERSION_NAME }}
              match:
              - headers:
                  branch:
                    exact: {{ (index $versions (add1 $i)).VERSION_NAME }}
              rewrite:
                uri: /v2/models/{{ (index $versions (add1 $i)).VERSION_NAME }}/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: {{ (index $versions (add1 $i)).VERSION_NAME }}-{{ template "kserve.host" $ }}
                    remove:
                    - branch
                  response:
                    add:
                      app-version: {{ (index $versions (add1 $i)).VERSION_NAME }}
              {{ `{{- end }}`}}     
              {{- end }}
            # traffic split
            - name: split
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                {{- if gt (len $versions) 1 }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}`}}
                {{- end  }}
                headers:
                  request:
                    set:
                      branch: {{ (index $versions 0).VERSION_NAME }}
                      host: {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                weight: {{ `{{ index .Weights `}}{{ print (add1 $i) }}{{` }}`}}
                headers:
                  request:
                    set:
                      branch: {{ (index $versions (add1 $i)).VERSION_NAME }}
                      host: {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
              {{ `{{- end }}`}}
              {{- end }}

{{- end }} {{- /* define "env.kserve.blue-green.routemap" */}}
