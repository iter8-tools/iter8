{{- define "initial.virtualservice" }}
{{- $versions := include "resolve.modelVersions" . | mustFromJson }}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Values.modelName }}
spec:
  gateways:
  - {{ .Values.externalGateway }}
  - mesh
  hosts:
  - {{ .Values.modelName }}.{{ .Release.Namespace }}
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc.cluster.local
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
{{- end }}
