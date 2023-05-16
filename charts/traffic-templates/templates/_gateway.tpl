{{- define "initial.gateway" }}
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: {{ .Values.externalGateway }}
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
{{- end }}