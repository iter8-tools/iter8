{{- define "env.kserve.blue-green.routemap" }}

{{- $APP_NAME := (include "application.name" .) }}
{{- $APP_NAMESPACE := (include "application.namespace" .) }}
{{- $versions := include "normalize.versions.kserve" . | mustFromJson }}

apiVersion: v1
kind: ConfigMap
{{- template "routemap.metadata" . }}
data:
  strSpec: |
    versions: 
    {{- range $i, $v := $versions }}
    - resources:
      - gvrShort: isvc
        name: {{ template "isvc.name" $v }}
        namespace: {{ template "isvc.namespace" $v }}
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
            - knative-serving/knative-ingress-gateway
            - knative-serving/knative-local-gateway
            - mesh
            hosts:
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc
            - {{ $APP_NAME }}.{{ $APP_NAMESPACE }}.svc.cluster.local
            http:
            # primary model version
            {{- $v := (index $versions 0) }}
            - name: {{ template "isvc.name" $v }}
              match:
              - headers:
                  branch:
                    exact: {{ template "isvc.name" $v }}
              rewrite:
                uri: /v2/models/{{ template "isvc.name" $v }}/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: {{ template "isvc.name" $v }}-{{ template "kserve.host" $ }}
                    remove:
                    - branch
                  response:
                    add:
                      app-version: {{ template "isvc.name" $v }}
            # other model versions
            {{- range $i, $v := (rest $versions) }}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}` }}
            - name: {{ template "isvc.name" $v }}
              match:
              - headers:
                  branch:
                    exact: {{ template "isvc.name" $v }}
              rewrite:
                uri: /v2/models/{{ template "isvc.name" $v }}/infer
              route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: {{ template "isvc.name" $v }}-{{ template "kserve.host" $ }}
                    remove:
                    - branch
                  response:
                    add:
                      app-version: {{ template "isvc.name" $v }}
              {{ `{{- end }}` }}     
              {{- end }}
            # traffic split
            - name: split
              route:
              {{- $v := (index $versions 0) }}
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                {{- if gt (len $versions) 1 }}
                {{ `{{- if gt (index .Weights 1) 0 }}` }}
                weight: {{ `{{ index .Weights 0 }}` }}
                {{ `{{- end }}` }}
                {{- end  }}
                headers:
                  request:
                    set:
                      branch: {{ template "isvc.name" $v }}
                      host: {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
              {{- range $i, $v := (rest $versions) }}
              {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}` }}
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                weight: {{ `{{ index .Weights ` }}{{ print (add1 $i) }}{{ ` }}` }}
                headers:
                  request:
                    set:
                      branch: {{ template "isvc.name" $v }}
                      host: {{ $APP_NAME }}.{{ $APP_NAMESPACE }}
              {{ `{{- end }}` }}
              {{- end }}

{{- end }} {{- /* define "env.kserve.blue-green.routemap" */}}
