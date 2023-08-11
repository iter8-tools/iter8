{{- define "mm.serviceentry" }}
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: {{ .Values.appName }}
spec:
  hosts:
  - {{ .Values.appName }}.{{ .Release.Namespace }}
  - {{ .Values.appName }}.{{ .Release.Namespace }}.svc
  - {{ .Values.appName }}.{{ .Release.Namespace }}.svc.cluster.local
  location: MESH_INTERNAL
  ports:
  - number: {{ .Values.modelmeshServingPort }}
    name: http
    protocol: HTTP
  resolution: DNS
  workloadSelector:
    labels:
      modelmesh-service: modelmesh-serving
{{- end }}