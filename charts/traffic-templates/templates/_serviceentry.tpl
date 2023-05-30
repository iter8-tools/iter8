{{- define "initial.serviceentry" }}
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: {{ .Values.modelName }}
spec:
  hosts:
  - {{ .Values.modelName }}.{{ .Release.Namespace }}
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc.cluster.local
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