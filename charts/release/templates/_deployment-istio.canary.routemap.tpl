{{- define "env.deployment-istio.canary.routemap" }}

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
    {{- end }}
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
            # non-primary versions
            {{- range $i, $v := (rest $versions) }}
            {{- /* continue only if candidate is ready (weight > 0) */}}
            {{ `{{- if gt (index .Weights ` }}{{ print (add1 $i) }}{{ `) 0 }}` }}
            - name: {{ template "svc.name" $v }}
              match:
              {{- /* A match may have several ORed clauses */}}
              {{- range $j, $m := $v.match }}
              {{- /* include any other header requirements */}}
              {{- if (hasKey $m "headers") }}
              - headers:
{{ toYaml (pick $m "headers").headers | indent 18 }}
              {{- end }} {{- /* if (hasKey $m "headers") */}}
              {{- /* include any other (non-header) requirements */}}
              {{- if gt (omit $m "headers" | keys | len) 0 }}
{{ toYaml (omit $m "headers") | indent 16 }}
              {{- end }} {{- /* if gt (omit $m "headers" | keys | len) 0 */}}
              {{- end }} {{- /* range $j, $m := $v.match */}}
              route:
              - destination:
                  host: {{ template "svc.name" $v }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                  port:
                    number: {{ $v.port }}
                headers: 
                  response:
                    add:
                      app-version: {{ template "svc.name" $v }}
              {{ `{{- end }}` }}
              {{- end }} {{- /* range $i, $v := (rest $versions) */}}
              # primary version (default)
              {{- $v := (index $versions 0) }}
              - name: {{ template "svc.name" $v }}
                route:
                - destination:
                    host: {{ template "svc.name" $v }}.{{ $APP_NAMESPACE }}.svc.cluster.local
                    port:
                      number: {{ $v.port }}
                  headers:
                    response:
                      add:
                        app-version: {{ template "svc.name" $v }}

{{- end }} {{- /* define "env.deployment-istio.canary.routemap" */}}
