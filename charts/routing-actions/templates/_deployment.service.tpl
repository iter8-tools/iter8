{{- define "deployment.service" }}
apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.appName }}
spec:
  externalName: istio-ingressgateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
{{- end }}