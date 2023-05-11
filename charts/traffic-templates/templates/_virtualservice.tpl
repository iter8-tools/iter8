{{- define "initial.virtualservice" }}
{{- $versions := include "resolve.modelVersions" . | mustFromJson }}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Values.modelName }}
spec:
  gateways:
  - {{ .Values.externalGateway }}
  hosts:
  - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}
  - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}.svc
  - {{ .Values.modelName }}.{{ .Values.modelmeshServingNamespace }}.svc.cluster.local
  http:
  - route:
    - destination:
        host: {{ .Values.modelmeshServingService }}.{{ .Values.modelmeshServingNamespace }}.svc.cluster.local
        port:
          number: {{ .Values.modelmeshServingPort }}
      headers:
        request:
          set:
            mm-vmodel-id: {{ (index $versions 0).name }}
{{- end }}
