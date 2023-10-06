{{- define "env.mm-istio.canary.routemap" }}

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
      - gvrShort: isvc
        name: {{ $v.VERSION_NAME }}
        namespace: {{ $v.VERSION_NAMESPACE }}
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
            # non-primary model versions

            {{- /* For candidate versions, ensure mm-model header is required in all matches */}}
            {{- range $i, $v := (rest $versions) }}
            {{- /* continue only if candidate is ready (ie, weight > 0) */}}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}`}}
            - match:
              {{- /* A match may have several ORed clauses */}}
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
                  host: {{ template "mm.serviceHost" $ }}
                  port:
                    number: {{ template "mm.servicePort" $ }}
                headers:
                  request:
                    set:
                      mm-vmodel-id: {{ (index $versions (add1 $i)).VERSION_NAME }}
                  response:
                    add:
                      app-version: {{ (index $versions (add1 $i)).VERSION_NAME }}
            {{ `{{- end }}`}}
            {{- end }}
            # primary model version (default)
            - name: {{ (index $versions 0).VERSION_NAME }}
              route:
              - destination:
                  host: {{ template "mm.serviceHost" }}
                  port:
                    number: {{ template "mm.servicePort" . }}
                headers:
                  request:
                    set:
                      mm-vmodel-id: {{ (index $versions 0).VERSION_NAME }}
                    remove:
                    - branch
                  response:
                    add:
                      app-version: {{ (index $versions 0).VERSION_NAME }}

{{- end }} {{- /* define "env.mm-istio.canary.routemap" */}}
