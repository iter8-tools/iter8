cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: wisdom
spec:
  externalName: knative-local-gateway.istio-system.svc.cluster.local
  sessionAffinity: None
  type: ExternalName
---
apiVersion: v1
kind: Namespace
metadata:
  name: primary
---
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: wisdom
  namespace: primary
  annotations:
    proxy.istio.io/config: '{"holdApplicationUntilProxyStarts": true}'
  labels:
    app.kubernetes.io/name: wisdom
    app.kubernetes.io/version: v1
    iter8.tools/watch: "true"
spec:
  predictor:
    model:
      modelFormat:
        name: sklearn
      runtime: kserve-mlserver
      storageUri: "gs://seldon-models/sklearn/mms/lr_model"
      ports:
      - containerPort: 9000
        name: h2c
        protocol: TCP      
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: wisdom
  labels:
    app.kubernetes.io/managed-by: iter8
    iter8.tools/kind: routemap
    iter8.tools/version: v0.14
data:
  strSpec: |
    variants: 
    - resources:
      - gvrShort: isvc
        name: wisdom
        namespace: primary
    - weight: 100
      resources:
      - gvrShort: isvc
        name: wisdom
        namespace: candidate
    # routing templates
    routingTemplates:
      wisdom:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: wisdom
          spec:
            gateways:
            - mesh
            hosts:
            - wisdom.default
            - wisdom.default.svc
            - wisdom.default.svc.cluster.local
            http:
            - route:
              - destination:
                  host: wisdom-predictor-default.primary.svc.cluster.local
              rewrite:
                authority: wisdom-predictor-default.primary.svc.cluster.local
              {{- if gt (index .Weights 1) 0 }}
              mirror:
                host: knative-local-gateway.istio-system.svc.cluster.local
              mirrorPercentage:
                value: {{ index .Weights 1 }}
              {{- end }}
immutable: true
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: wisdom-mirror
spec:
  gateways:
  - knative-serving/knative-local-gateway
  hosts:
  - "*"
  http:
  - match:
    - authority:
        prefix: wisdom-predictor-default.primary.svc.cluster.local-shadow
    route:
    - destination:
        host: wisdom-predictor-default.candidate.svc.cluster.local
    rewrite:
      authority: wisdom-predictor-default.candidate.svc.cluster.local
EOF