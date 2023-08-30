{{- define "kserve.virtualservice-canary" }}
{{- $versions := include "resolve.appVersions" . | mustFromJson }}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Values.appName }}
spec:
  gateways:
  - knative-serving/knative-ingress-gateway
  - knative-serving/knative-local-gateway
  - mesh
  hosts:
  - {{ .Values.appName }}.{{ .Release.Namespace }}
  - {{ .Values.appName }}.{{ .Release.Namespace }}.svc
  - {{ .Values.appName }}.{{ .Release.Namespace }}.svc.cluster.local
  http:
  - name: {{ (index $versions 0).name }}
    rewrite:
      uri: /v2/models/{{ (index $versions 0).name }}/infer
    route:
    - destination:
        host: knative-local-gateway.istio-system.svc.cluster.local
      headers: 
        request:
          set:
            Host: {{ (index $versions 0).name }}-{{ template "kserve.host" $ }}
        response:
          add:
            app-version: {{ (index $versions 0).name }}
{{- end }}
