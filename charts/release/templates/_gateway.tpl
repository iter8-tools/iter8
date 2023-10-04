{{- define "default.gateway" }}
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: {{ default "iter8-gateway" .Values.gateway }}
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
{{- end }} {{- /* define "default.gateway" */}}