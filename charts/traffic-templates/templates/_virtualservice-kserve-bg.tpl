{{- define "initial.virtualservice-kserve-bg" }}
{{- $versions := include "resolve.modelVersions" . | mustFromJson }}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Values.modelName }}
spec:
  gateways:
  {{- if eq "kserve" .Values.targetEnv }}
  - knative-serving/knative-ingress-gateway
  - knative-serving/knative-local-gateway
  {{- else if eq "kserve-modelmesh" .Values.targetEnv }}
  - {{ .Values.externalGateway }}
  {{- end }}
  - mesh
  hosts:
  - {{ .Values.modelName }}.{{ .Release.Namespace }}
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc
  - {{ .Values.modelName }}.{{ .Release.Namespace }}.svc.cluster.local
  http:
  - name: {{ (index $versions 0).name }}
    match:
    - headers:
        branch:
          exact: {{ (index $versions 0).name }}
    rewrite:
      uri: /v2/models/{{ (index $versions 0).name }}/infer
    route:
    # primary model
    - destination:
        host: knative-local-gateway.istio-system.svc.cluster.local
      headers: 
        request:
          set:
            Host: {{ (index $versions 0).name }}-predictor-default.{{ .Release.Namespace }}.svc.cluster.local
          remove: 
          - branch
        response:
          add:
            mm-vmodel-id: {{ (index $versions 0).name }}
  - name: split
    route:
    - destination:
        host: knative-local-gateway.istio-system.svc.cluster.local
      headers:
        request:
          set:
            branch: {{ (index $versions 0).name }}
            host: {{ .Values.modelName }}.{{ .Release.Namespace }}
{{- end }}
