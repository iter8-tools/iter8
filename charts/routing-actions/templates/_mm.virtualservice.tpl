{{- define "mm.virtualservice" }}
{{- $versions := include "resolve.appVersions" . | mustFromJson }}
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
          number: {{ .Values.modelmeshServingPort }}
      headers:
        request:
          set:
            mm-vmodel-id: {{ (index $versions 0).name }}
        response:
          add:
            app-version: "{{ (index $versions 0).name }}"
{{- end }}
