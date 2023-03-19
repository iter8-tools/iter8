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
    iter8.tools/kind: subject
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
    ssas:
      mirror-wisdom:
        gvrShort: vs
        template: |
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          metadata:
            name: wisdom
          spec:
            gateways:
            - mesh
            - knative-serving/knative-ingress-gateway
            - knative-serving/knative-local-gateway
            hosts:
            - wisdom.default
            - wisdom.default.svc.cluster.local
            http:
            - route:
              - destination:
                  host: knative-local-gateway.istio-system.svc.cluster.local
                headers:
                  request:
                    set:
                      Host: wisdom-predictor-default.primary.svc.cluster.local
              {{- if gt (index .Weights 1) 0 }}
              mirror:
                host: wisdom-predictor-default-00001.candidate
              mirrorPercentage:
                value: {{ index .Weights 1 }}
              {{- end }}
immutable: true            
EOF