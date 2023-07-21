{{- define "initial.service-kserve" }}
apiVersion: v1
kind: Service
metadata:
  name: {{ .Values.modelName }}
spec:
  externalName: knative-local-gateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
{{- end }}